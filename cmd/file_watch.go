package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail <file|profile:/path|[user@]host:/path>",
	Short: "Follow local or remote log output",
	Long: `Print the end of a file and optionally follow changes.

Remote paths use SSH: profile:/path or [user@]host:/path. Remote auth uses the same auto key
bootstrap behavior as acli ssh.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")
		if err := runTailTarget(cmd, args[0], lines, follow); err != nil {
			exitErr(err)
		}
	},
}

var watchfileCmd = &cobra.Command{
	Use:   "watchfile <path>",
	Short: "Watch a file or directory for changes",
	Long:  `Poll a file or directory and print create, modify, and delete events. Works on Windows, Linux, and macOS without platform-specific watchers.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetDuration("interval")
		if interval <= 0 {
			interval = time.Second
		}
		if err := watchPath(args[0], interval); err != nil {
			exitErr(err)
		}
	},
}

func init() {
	tailCmd.Flags().IntP("lines", "n", 10, "Number of lines to show")
	tailCmd.Flags().BoolP("follow", "f", false, "Follow appended data")
	for _, c := range []*cobra.Command{tailCmd} {
		c.Flags().StringP("user", "u", "", "SSH username for remote path")
		c.Flags().IntP("port", "p", defaultSSHPort, "SSH port for remote path")
		c.Flags().StringP("key", "i", "", "Private key path for remote path")
		c.Flags().String("passphrase", "", "Private key passphrase for remote path")
		c.Flags().String("password", "", "SSH password for remote path")
		c.Flags().Bool("no-bootstrap", false, "Do not install default public key after password login")
		c.Flags().Duration("timeout", 15*time.Second, "SSH connection timeout")
	}
	watchfileCmd.Flags().DurationP("interval", "i", time.Second, "Polling interval")
}

func runTailTarget(cmd *cobra.Command, target string, lines int, follow bool) error {
	if remote := parseRemotePath(target); remote != nil {
		opts, err := sshOptionsFromFlags(cmd, remote.host)
		if err != nil {
			return err
		}
		if remote.user != "" {
			opts.user = remote.user
		}
		client, usedPassword, generatedPub, err := connectSSH(opts)
		if err != nil {
			return err
		}
		defer client.Close()
		if usedPassword && !opts.noBootstrap {
			if err := installRemoteAuthorizedKey(client, generatedPub); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: key bootstrap failed: %v\n", err)
			}
		}

		flag := "-n"
		if follow {
			flag = "-f -n"
		}
		return runSSHCommand(client, fmt.Sprintf("tail %s %d %s", flag, lines, shellQuote(remote.path)), nil, os.Stdout, os.Stderr)
	}
	return tailLocal(target, lines, follow)
}

func tailLocal(path string, lines int, follow bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	offset, err := printLastLines(file, lines)
	if err != nil {
		return err
	}
	if !follow {
		return nil
	}

	for {
		time.Sleep(time.Second)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Size() < offset {
			offset = 0
		}
		if info.Size() == offset {
			continue
		}
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		copied, err := io.Copy(os.Stdout, file)
		if err != nil {
			return err
		}
		offset += copied
	}
}

func printLastLines(file *os.File, lines int) (int64, error) {
	if lines < 0 {
		lines = 0
	}
	scanner := bufio.NewScanner(file)
	buf := make([]string, 0, lines)
	for scanner.Scan() {
		if lines == 0 {
			continue
		}
		if len(buf) == lines {
			copy(buf, buf[1:])
			buf[len(buf)-1] = scanner.Text()
		} else {
			buf = append(buf, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	for _, line := range buf {
		fmt.Println(line)
	}
	return file.Seek(0, io.SeekEnd)
}

type fileSnapshot struct {
	size    int64
	modTime time.Time
	isDir   bool
}

func watchPath(path string, interval time.Duration) error {
	previous, err := snapshotPath(path)
	if err != nil {
		return err
	}
	fmt.Printf("Watching %s every %s\n", path, interval)
	for {
		time.Sleep(interval)
		current, err := snapshotPath(path)
		if err != nil {
			if len(previous) > 0 {
				for name := range previous {
					fmt.Printf("DELETE %s\n", name)
				}
				previous = map[string]fileSnapshot{}
				continue
			}
			return err
		}
		for name, cur := range current {
			prev, ok := previous[name]
			if !ok {
				fmt.Printf("CREATE %s\n", name)
				continue
			}
			if prev.size != cur.size || !prev.modTime.Equal(cur.modTime) || prev.isDir != cur.isDir {
				fmt.Printf("MODIFY %s\n", name)
			}
		}
		for name := range previous {
			if _, ok := current[name]; !ok {
				fmt.Printf("DELETE %s\n", name)
			}
		}
		previous = current
	}
}

func snapshotPath(path string) (map[string]fileSnapshot, error) {
	result := map[string]fileSnapshot{}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		result[path] = fileSnapshot{size: info.Size(), modTime: info.ModTime(), isDir: false}
		return result, nil
	}
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel := p
		if r, err := filepath.Rel(path, p); err == nil && r != "." {
			rel = strings.ReplaceAll(r, string(os.PathSeparator), "/")
		}
		result[rel] = fileSnapshot{size: info.Size(), modTime: info.ModTime(), isDir: info.IsDir()}
		return nil
	})
	return result, err
}
