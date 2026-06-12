package cmd

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

const defaultSSHPort = 22

type sshOptions struct {
	name        string
	user        string
	host        string
	port        int
	keyPath     string
	passphrase  string
	password    string
	noBootstrap bool
	timeout     time.Duration
}

type sshProfile struct {
	Name    string   `json:"name"`
	Host    string   `json:"host"`
	User    string   `json:"user"`
	Port    int      `json:"port"`
	KeyPath string   `json:"key_path,omitempty"`
	Groups  []string `json:"groups,omitempty"`
}

var sshCmd = &cobra.Command{
	Use:   "ssh [profile|user@host] [command...]",
	Short: "SSH command, shell, and account management",
	Long: `Run SSH commands or open an interactive shell.

Auth order: default ~/.ssh keys, --key/profile key, then password. If password
auth succeeds, unicli appends the default public key to remote authorized_keys.
Passwords are never saved in ssh profiles. Command stdin/stdout/stderr are streamed
live; unicli does not wait for the remote command to finish before printing output.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := sshOptionsFromFlags(cmd, args[0])
		if err != nil {
			exitErr(err)
		}

		remoteCmd := strings.Join(args[1:], " ")
		client, usedPassword, bootstrapPub, err := connectSSH(opts)
		if err != nil {
			exitErr(err)
		}
		defer client.Close()

		if usedPassword && !opts.noBootstrap && len(bootstrapPub) > 0 {
			if err := installRemoteAuthorizedKey(client, bootstrapPub); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: key bootstrap failed: %v\n", err)
			} else {
				fmt.Println("Key installed: future logins can use default SSH key")
			}
		}

		if remoteCmd == "" {
			if err := runInteractiveSSH(client); err != nil {
				exitErr(err)
			}
			return
		}
		if err := runSSHCommand(client, remoteCmd, os.Stdin, os.Stdout, os.Stderr); err != nil {
			exitErr(err)
		}
	},
}

var sshAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update SSH account profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		keyPath, _ := cmd.Flags().GetString("key")
		groups, _ := cmd.Flags().GetStringSlice("group")
		if host == "" || user == "" {
			exitErr(errors.New("--host and --user are required"))
		}
		if port <= 0 {
			port = defaultSSHPort
		}
		profiles, err := loadSSHProfiles()
		if err != nil {
			exitErr(err)
		}
		profiles[args[0]] = sshProfile{Name: args[0], Host: host, User: user, Port: port, KeyPath: keyPath, Groups: groups}
		if err := saveSSHProfiles(profiles); err != nil {
			exitErr(err)
		}
		fmt.Printf("Saved SSH profile: %s\n", args[0])
	},
}

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSH account profiles",
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := loadSSHProfiles()
		if err != nil {
			exitErr(err)
		}
		names := make([]string, 0, len(profiles))
		for name := range profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Printf("%-18s %-24s %-18s %-6s %-18s %s\n", "NAME", "HOST", "USER", "PORT", "GROUPS", "KEY")
		for _, name := range names {
			p := profiles[name]
			fmt.Printf("%-18s %-24s %-18s %-6d %-18s %s\n", p.Name, p.Host, p.User, p.Port, strings.Join(p.Groups, ","), p.KeyPath)
		}
	},
}

var sshExecCmd = &cobra.Command{
	Use:   "exec (--all|--group name|profile...) -- <command>",
	Short: "Run command on multiple SSH profiles with live prefixed output",
	Long:  `Run a command across SSH profiles. Output streams live with [profile] prefixes for agent-readable remote operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		group, _ := cmd.Flags().GetString("group")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		if len(args) == 0 {
			exitErr(errors.New("missing command"))
		}
		profileNames, remoteArgs := splitProfilesAndCommand(args)
		if len(remoteArgs) == 0 {
			exitErr(errors.New("missing command"))
		}
		targets, err := selectSSHProfiles(profileNames, all, group)
		if err != nil {
			exitErr(err)
		}
		remoteCmd := strings.Join(remoteArgs, " ")
		exitCode := 0
		for _, profile := range targets {
			fmt.Printf("[%s] START %s\n", profile.Name, remoteCmd)
			opts := sshOptions{name: profile.Name, user: profile.User, host: profile.Host, port: profile.Port, keyPath: profile.KeyPath, timeout: timeout}
			client, usedPassword, bootstrapPub, err := connectSSH(opts)
			if err != nil {
				fmt.Printf("[%s] ERROR connect: %v\n", profile.Name, err)
				exitCode = 1
				continue
			}
			if usedPassword && len(bootstrapPub) > 0 {
				_ = installRemoteAuthorizedKey(client, bootstrapPub)
			}
			err = runSSHCommand(client, remoteCmd, nil, &prefixWriter{prefix: "[" + profile.Name + "] ", writer: os.Stdout}, &prefixWriter{prefix: "[" + profile.Name + "] ERR ", writer: os.Stderr})
			client.Close()
			if err != nil {
				fmt.Printf("[%s] EXIT error=%v\n", profile.Name, err)
				exitCode = 1
			} else {
				fmt.Printf("[%s] OK\n", profile.Name)
			}
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	},
}

var sshRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove SSH account profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := loadSSHProfiles()
		if err != nil {
			exitErr(err)
		}
		delete(profiles, args[0])
		if err := saveSSHProfiles(profiles); err != nil {
			exitErr(err)
		}
		fmt.Printf("Removed SSH profile: %s\n", args[0])
	},
}

var sshExportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Export SSH account profiles without passwords",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := loadSSHProfiles()
		if err != nil {
			exitErr(err)
		}
		data, err := json.MarshalIndent(profiles, "", "  ")
		if err != nil {
			exitErr(err)
		}
		if err := os.WriteFile(args[0], data, 0600); err != nil {
			exitErr(err)
		}
		fmt.Printf("Exported SSH profiles: %s\n", args[0])
	},
}

var sshImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import SSH account profiles",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			exitErr(err)
		}
		imported := map[string]sshProfile{}
		if err := json.Unmarshal(data, &imported); err != nil {
			exitErr(err)
		}
		profiles, err := loadSSHProfiles()
		if err != nil {
			exitErr(err)
		}
		for name, p := range imported {
			if p.Name == "" {
				p.Name = name
			}
			if p.Port <= 0 {
				p.Port = defaultSSHPort
			}
			profiles[p.Name] = p
		}
		if err := saveSSHProfiles(profiles); err != nil {
			exitErr(err)
		}
		fmt.Printf("Imported %d SSH profiles\n", len(imported))
	},
}

var scpCmd = &cobra.Command{
	Use:   "scp <source> <destination>",
	Short: "Copy files over SSH/SFTP",
	Long: `Copy files or directories over SSH/SFTP.

Remote paths use [profile:]path or [user@]host:/path. Examples:
  unicli scp ./app.log prod:/tmp/app.log
  unicli scp user@example.com:/var/log/app.log ./app.log

Transfers stream bytes through SFTP; files are not buffered fully in memory.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		src, dst := args[0], args[1]
		recursive, _ := cmd.Flags().GetBool("recursive")
		srcRemote := parseRemotePath(src)
		dstRemote := parseRemotePath(dst)
		if (srcRemote != nil && dstRemote != nil) || (srcRemote == nil && dstRemote == nil) {
			exitErr(errors.New("exactly one path must be remote: profile:/path or [user@]host:/path"))
		}

		target := srcRemote
		if target == nil {
			target = dstRemote
		}

		opts, err := sshOptionsFromFlags(cmd, target.host)
		if err != nil {
			exitErr(err)
		}
		if target.user != "" {
			opts.user = target.user
		}

		client, usedPassword, bootstrapPub, err := connectSSH(opts)
		if err != nil {
			exitErr(err)
		}
		defer client.Close()
		if usedPassword && !opts.noBootstrap && len(bootstrapPub) > 0 {
			if err := installRemoteAuthorizedKey(client, bootstrapPub); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: key bootstrap failed: %v\n", err)
			}
		}

		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			exitErr(err)
		}
		defer sftpClient.Close()

		if srcRemote != nil {
			if info, err := sftpClient.Stat(srcRemote.path); err == nil && info.IsDir() {
				if !recursive {
					exitErr(errors.New("remote source is a directory; pass --recursive"))
				}
				if err := syncRemoteToLocal(sftpClient, srcRemote.path, dst, false, false); err != nil {
					exitErr(err)
				}
				return
			}
			if err := sftpDownload(sftpClient, srcRemote.path, dst); err != nil {
				exitErr(err)
			}
			fmt.Printf("Downloaded %s -> %s\n", src, dst)
			return
		}
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			if !recursive {
				exitErr(errors.New("local source is a directory; pass --recursive"))
			}
			if err := syncLocalToRemote(sftpClient, src, dstRemote.path, false, false); err != nil {
				exitErr(err)
			}
			return
		}
		if err := sftpUpload(sftpClient, src, dstRemote.path); err != nil {
			exitErr(err)
		}
		fmt.Printf("Uploaded %s -> %s\n", src, dst)
	},
}

type remotePath struct {
	user string
	host string
	path string
}

func init() {
	for _, c := range []*cobra.Command{sshCmd, scpCmd} {
		c.Flags().StringP("user", "u", "", "SSH username")
		c.Flags().IntP("port", "p", defaultSSHPort, "SSH port")
		c.Flags().StringP("key", "i", "", "Private key path")
		c.Flags().String("passphrase", "", "Private key passphrase (prompted if needed and omitted)")
		c.Flags().String("password", "", "SSH password (prompted if needed and omitted)")
		c.Flags().Bool("no-bootstrap", false, "Do not install default public key after password login")
		c.Flags().Duration("timeout", 15*time.Second, "Connection timeout")
	}
	scpCmd.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	sshAddCmd.Flags().String("host", "", "SSH host")
	sshAddCmd.Flags().StringP("user", "u", "", "SSH username")
	sshAddCmd.Flags().IntP("port", "p", defaultSSHPort, "SSH port")
	sshAddCmd.Flags().StringP("key", "i", "", "Private key path")
	sshAddCmd.Flags().StringSlice("group", nil, "Profile group name (repeatable)")
	sshExecCmd.Flags().Bool("all", false, "Run on all profiles")
	sshExecCmd.Flags().String("group", "", "Run on profiles in group")
	sshExecCmd.Flags().Duration("timeout", 15*time.Second, "Connection timeout")
	sshCmd.AddCommand(sshAddCmd, sshListCmd, sshRemoveCmd, sshExportCmd, sshImportCmd, sshExecCmd)
}

func sshOptionsFromFlags(cmd *cobra.Command, target string) (sshOptions, error) {
	user, _ := cmd.Flags().GetString("user")
	port, _ := cmd.Flags().GetInt("port")
	keyPath, _ := cmd.Flags().GetString("key")
	passphrase, _ := cmd.Flags().GetString("passphrase")
	password, _ := cmd.Flags().GetString("password")
	noBootstrap, _ := cmd.Flags().GetBool("no-bootstrap")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	profiles, _ := loadSSHProfiles()
	if profile, ok := profiles[target]; ok {
		if user == "" {
			user = profile.User
		}
		if port == defaultSSHPort && profile.Port > 0 {
			port = profile.Port
		}
		if keyPath == "" {
			keyPath = profile.KeyPath
		}
		target = profile.Host
	}

	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		if user == "" {
			user = parts[0]
		}
		target = parts[1]
	}
	if strings.Contains(target, ":") {
		host, portText, err := net.SplitHostPort(target)
		if err == nil {
			target = host
			if parsedPort, parseErr := strconv.Atoi(portText); parseErr == nil {
				port = parsedPort
			}
		}
	}
	if user == "" {
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME")
		}
	}
	if user == "" {
		return sshOptions{}, errors.New("missing SSH user; pass profile, user@host, or --user")
	}
	if target == "" {
		return sshOptions{}, errors.New("missing SSH host")
	}
	if port <= 0 {
		port = defaultSSHPort
	}
	return sshOptions{name: target, user: user, host: target, port: port, keyPath: keyPath, passphrase: passphrase, password: password, noBootstrap: noBootstrap, timeout: timeout}, nil
}

func connectSSH(opts sshOptions) (*ssh.Client, bool, []byte, error) {
	keyPaths, bootstrapKey, err := defaultSSHKeyPaths()
	if err != nil {
		return nil, false, nil, err
	}
	if opts.keyPath != "" {
		keyPaths = append([]string{opts.keyPath}, keyPaths...)
	}

	authMethods := []ssh.AuthMethod{}
	askedPassphrase := false
	for _, keyPath := range keyPaths {
		signer, err := signerFromFileWithPassphrase(keyPath, opts.passphrase)
		if isPassphraseMissing(err) && !askedPassphrase {
			askedPassphrase = true
			fmt.Fprintf(os.Stderr, "Passphrase for %s: ", keyPath)
			pw, readErr := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if readErr == nil {
				opts.passphrase = string(pw)
				signer, err = signerFromFileWithPassphrase(keyPath, opts.passphrase)
			}
		}
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	var bootstrapPub []byte
	if bootstrapKey != "" {
		bootstrapPub, _ = publicKeyFromPrivateKeyFile(bootstrapKey)
	}
	if len(authMethods) > 0 {
		if client, err := dialSSH(opts, authMethods); err == nil {
			return client, false, bootstrapPub, nil
		}
	}

	password := opts.password
	if password == "" {
		fmt.Fprintf(os.Stderr, "Password for %s@%s: ", opts.user, net.JoinHostPort(opts.host, strconv.Itoa(opts.port)))
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return nil, false, nil, err
		}
		password = string(pw)
	}
	client, err := dialSSH(opts, []ssh.AuthMethod{ssh.Password(password)})
	if err != nil {
		return nil, false, nil, err
	}
	return client, true, bootstrapPub, nil
}

func dialSSH(opts sshOptions, auth []ssh.AuthMethod) (*ssh.Client, error) {
	hostKeyCallback, err := knownHostsCallback()
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            opts.user,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         opts.timeout,
	}
	return ssh.Dial("tcp", net.JoinHostPort(opts.host, strconv.Itoa(opts.port)), config)
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	path, err := knownHostsPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, nil, 0600); err != nil {
			return nil, err
		}
	}
	base, err := knownhosts.New(path)
	if err != nil {
		return nil, err
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := base(hostname, remote, key)
		if err == nil {
			return nil
		}
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			line := knownhosts.Line([]string{hostname}, key)
			f, openErr := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
			if openErr != nil {
				return openErr
			}
			defer f.Close()
			_, writeErr := fmt.Fprintln(f, line)
			return writeErr
		}
		return err
	}, nil
}

func runInteractiveSSH(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, rawErr := term.MakeRaw(fd)
	if rawErr == nil {
		defer term.Restore(fd, oldState)
	}
	width, height, sizeErr := term.GetSize(fd)
	if sizeErr != nil {
		width, height = 120, 40
	}
	terminalModes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := session.RequestPty(termName(), height, width, terminalModes); err != nil {
		return err
	}
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if err := session.Shell(); err != nil {
		return err
	}
	return session.Wait()
}

func termName() string {
	if v := os.Getenv("TERM"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		return "xterm"
	}
	return "vt100"
}

func defaultSSHKeyPaths() ([]string, string, error) {
	dir, err := userSSHDir()
	if err != nil {
		return nil, "", err
	}
	candidates := []string{
		filepath.Join(dir, "id_ed25519"),
		filepath.Join(dir, "id_ecdsa"),
		filepath.Join(dir, "id_rsa"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return candidates, path, nil
		}
	}
	keyPath := filepath.Join(dir, "id_ed25519")
	if err := generateEd25519Key(keyPath); err != nil {
		return nil, "", err
	}
	fmt.Fprintf(os.Stderr, "Generated SSH key: %s\n", keyPath)
	return candidates, keyPath, nil
}

func generateEd25519Key(keyPath string) error {
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(keyPath, pemBytes, 0600); err != nil {
		return err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return err
	}
	return os.WriteFile(keyPath+".pub", ssh.MarshalAuthorizedKey(sshPub), 0644)
}

func userSSHDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

func knownHostsPath() (string, error) {
	dir, err := userSSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "known_hosts"), nil
}

func sshProfilesPath() (string, error) {
	dir, err := userSSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "unicli_accounts.json"), nil
}

func loadSSHProfiles() (map[string]sshProfile, error) {
	path, err := sshProfilesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]sshProfile{}, nil
	}
	if err != nil {
		return nil, err
	}
	profiles := map[string]sshProfile{}
	if len(bytes.TrimSpace(data)) == 0 {
		return profiles, nil
	}
	return profiles, json.Unmarshal(data, &profiles)
}

func saveSSHProfiles(profiles map[string]sshProfile) error {
	path, err := sshProfilesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func splitProfilesAndCommand(args []string) ([]string, []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return nil, args
}

func selectSSHProfiles(names []string, all bool, group string) ([]sshProfile, error) {
	profiles, err := loadSSHProfiles()
	if err != nil {
		return nil, err
	}
	selected := []sshProfile{}
	if all || group != "" {
		keys := make([]string, 0, len(profiles))
		for name := range profiles {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			profile := profiles[name]
			if all || profileInGroup(profile, group) {
				selected = append(selected, profile)
			}
		}
	} else {
		for _, name := range names {
			profile, ok := profiles[name]
			if !ok {
				return nil, fmt.Errorf("unknown SSH profile: %s", name)
			}
			selected = append(selected, profile)
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("no SSH profiles selected")
	}
	return selected, nil
}

func profileInGroup(profile sshProfile, group string) bool {
	for _, item := range profile.Groups {
		if item == group {
			return true
		}
	}
	return false
}

type prefixWriter struct {
	prefix string
	writer io.Writer
	buffer []byte
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.buffer = append(w.buffer, p...)
	for {
		idx := bytes.IndexByte(w.buffer, '\n')
		if idx < 0 {
			break
		}
		line := string(w.buffer[:idx+1])
		if _, err := fmt.Fprint(w.writer, w.prefix+line); err != nil {
			return 0, err
		}
		w.buffer = w.buffer[idx+1:]
	}
	if len(w.buffer) > 0 && !bytes.Contains(w.buffer, []byte("\n")) {
		line := string(w.buffer)
		if _, err := fmt.Fprint(w.writer, w.prefix+line); err != nil {
			return 0, err
		}
		w.buffer = nil
	}
	return len(p), nil
}

func signerFromFile(path string) (ssh.Signer, error) {
	return signerFromFileWithPassphrase(path, "")
}

func signerFromFileWithPassphrase(path, passphrase string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(expandHome(path))
	if err != nil {
		return nil, err
	}
	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(keyBytes)
}

func isPassphraseMissing(err error) bool {
	var missing *ssh.PassphraseMissingError
	return errors.As(err, &missing)
}

func publicKeyFromPrivateKeyFile(path string) ([]byte, error) {
	signer, err := signerFromFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(signer.PublicKey()), nil
}

func installRemoteAuthorizedKey(client *ssh.Client, publicKey []byte) error {
	escapedKey := shellQuote(strings.TrimSpace(string(publicKey)))
	script := "mkdir -p ~/.ssh && chmod 700 ~/.ssh && touch ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && grep -qxF " + escapedKey + " ~/.ssh/authorized_keys || printf '%s\\n' " + escapedKey + " >> ~/.ssh/authorized_keys"
	var stderr bytes.Buffer
	if err := runSSHCommand(client, script, nil, io.Discard, &stderr); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func runSSHCommand(client *ssh.Client, remoteCmd string, stdin io.Reader, stdout, stderr io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	if stdin != nil {
		session.Stdin = stdin
	}
	session.Stdout = stdout
	session.Stderr = stderr
	return session.Run(remoteCmd)
}

func parseRemotePath(value string) *remotePath {
	if filepath.VolumeName(value) != "" {
		return nil
	}
	idx := strings.Index(value, ":")
	if idx <= 0 {
		return nil
	}
	left := value[:idx]
	right := value[idx+1:]
	if right == "" || strings.Contains(left, string(os.PathSeparator)) {
		return nil
	}
	r := &remotePath{host: left, path: right}
	if strings.Contains(left, "@") {
		parts := strings.SplitN(left, "@", 2)
		r.user = parts[0]
		r.host = parts[1]
	}
	return r
}

func sftpDownload(client *sftp.Client, remoteFile, localFile string) error {
	src, err := client.Open(remoteFile)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func sftpUpload(client *sftp.Client, localFile, remoteFile string) error {
	src, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := client.Create(remoteFile)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func expandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
