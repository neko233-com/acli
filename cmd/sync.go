package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <source> <destination>",
	Short: "Sync files or directories over SSH/SFTP",
	Long:  `Lightweight rsync-like sync. Exactly one side must be remote: profile:/path or [user@]host:/path. Files stream through SFTP and are copied when size or modtime differs.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		deleteExtra, _ := cmd.Flags().GetBool("delete")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if err := runSync(cmd, args[0], args[1], deleteExtra, dryRun); err != nil {
			exitErr(err)
		}
	},
}

func init() {
	syncCmd.Flags().Bool("delete", false, "Delete destination files missing from source")
	syncCmd.Flags().Bool("dry-run", false, "Print planned changes without copying")
	for _, c := range []*cobra.Command{syncCmd} {
		c.Flags().StringP("user", "u", "", "SSH username")
		c.Flags().IntP("port", "p", defaultSSHPort, "SSH port")
		c.Flags().StringP("key", "i", "", "Private key path")
		c.Flags().String("passphrase", "", "Private key passphrase")
		c.Flags().String("password", "", "SSH password")
		c.Flags().Bool("no-bootstrap", false, "Do not install default public key after password login")
		c.Flags().Duration("timeout", 15*time.Second, "Connection timeout")
	}
}

func runSync(cmd *cobra.Command, src, dst string, deleteExtra, dryRun bool) error {
	srcRemote := parseRemotePath(src)
	dstRemote := parseRemotePath(dst)
	if (srcRemote != nil && dstRemote != nil) || (srcRemote == nil && dstRemote == nil) {
		return errors.New("exactly one path must be remote: profile:/path or [user@]host:/path")
	}
	target := srcRemote
	if target == nil {
		target = dstRemote
	}
	client, sftpClient, err := connectRemotePath(cmd, target)
	if err != nil {
		return err
	}
	defer client.Close()
	defer sftpClient.Close()

	if srcRemote != nil {
		return syncRemoteToLocal(sftpClient, srcRemote.path, dst, deleteExtra, dryRun)
	}
	return syncLocalToRemote(sftpClient, src, dstRemote.path, deleteExtra, dryRun)
}

func connectRemotePath(cmd *cobra.Command, remote *remotePath) (*sshClientCloser, *sftp.Client, error) {
	opts, err := sshOptionsFromFlags(cmd, remote.host)
	if err != nil {
		return nil, nil, err
	}
	if remote.user != "" {
		opts.user = remote.user
	}
	client, usedPassword, bootstrapPub, err := connectSSH(opts)
	if err != nil {
		return nil, nil, err
	}
	if usedPassword && !opts.noBootstrap && len(bootstrapPub) > 0 {
		_ = installRemoteAuthorizedKey(client, bootstrapPub)
	}
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return &sshClientCloser{close: client.Close}, sftpClient, nil
}

type sshClientCloser struct {
	close func() error
}

func (c *sshClientCloser) Close() error {
	if c == nil || c.close == nil {
		return nil
	}
	return c.close()
}

func syncLocalToRemote(client *sftp.Client, localRoot, remoteRoot string, deleteExtra, dryRun bool) error {
	source := map[string]os.FileInfo{}
	err := filepath.Walk(localRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(localRoot, path)
		if rel == "." {
			return nil
		}
		source[filepath.ToSlash(rel)] = info
		remotePath := joinRemote(remoteRoot, filepath.ToSlash(rel))
		if info.IsDir() {
			if dryRun {
				fmt.Printf("MKDIR %s\n", remotePath)
				return nil
			}
			return client.MkdirAll(remotePath)
		}
		remoteInfo, statErr := client.Stat(remotePath)
		if statErr == nil && sameFile(info, remoteInfo) {
			return nil
		}
		fmt.Printf("COPY %s -> %s\n", path, remotePath)
		if dryRun {
			return nil
		}
		return uploadFile(client, path, remotePath, info.ModTime())
	})
	if err != nil {
		return err
	}
	if deleteExtra {
		return deleteRemoteExtras(client, remoteRoot, source, dryRun)
	}
	return nil
}

func syncRemoteToLocal(client *sftp.Client, remoteRoot, localRoot string, deleteExtra, dryRun bool) error {
	source := map[string]os.FileInfo{}
	walker := client.Walk(remoteRoot)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		path := walker.Path()
		info := walker.Stat()
		rel := strings.TrimPrefix(strings.TrimPrefix(path, remoteRoot), "/")
		if rel == "" {
			continue
		}
		source[rel] = info
		localPath := filepath.Join(localRoot, filepath.FromSlash(rel))
		if info.IsDir() {
			if dryRun {
				fmt.Printf("MKDIR %s\n", localPath)
				continue
			}
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return err
			}
			continue
		}
		localInfo, statErr := os.Stat(localPath)
		if statErr == nil && sameFile(info, localInfo) {
			continue
		}
		fmt.Printf("COPY %s -> %s\n", path, localPath)
		if dryRun {
			continue
		}
		if err := downloadFile(client, path, localPath, info.ModTime()); err != nil {
			return err
		}
	}
	if deleteExtra {
		return deleteLocalExtras(localRoot, source, dryRun)
	}
	return nil
}

func uploadFile(client *sftp.Client, localPath, remotePath string, modTime time.Time) error {
	if err := client.MkdirAll(pathDir(remotePath)); err != nil {
		return err
	}
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := client.Create(remotePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return client.Chtimes(remotePath, modTime, modTime)
}

func downloadFile(client *sftp.Client, remotePath, localPath string, modTime time.Time) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}
	src, err := client.Open(remotePath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Chtimes(localPath, modTime, modTime)
}

func sameFile(a, b os.FileInfo) bool {
	return a.Size() == b.Size() && a.ModTime().Unix() == b.ModTime().Unix()
}

func joinRemote(root, rel string) string {
	return strings.TrimRight(root, "/") + "/" + strings.TrimLeft(rel, "/")
}

func pathDir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return "."
	}
	return path[:idx]
}

func deleteRemoteExtras(client *sftp.Client, remoteRoot string, source map[string]os.FileInfo, dryRun bool) error {
	walker := client.Walk(remoteRoot)
	paths := []string{}
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(walker.Path(), remoteRoot), "/")
		if rel != "" {
			paths = append(paths, rel)
		}
	}
	for i := len(paths) - 1; i >= 0; i-- {
		rel := paths[i]
		if _, ok := source[rel]; ok {
			continue
		}
		target := joinRemote(remoteRoot, rel)
		fmt.Printf("DELETE %s\n", target)
		if dryRun {
			continue
		}
		_ = client.Remove(target)
		_ = client.RemoveDirectory(target)
	}
	return nil
}

func deleteLocalExtras(localRoot string, source map[string]os.FileInfo, dryRun bool) error {
	paths := []string{}
	if err := filepath.Walk(localRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(localRoot, path)
		if rel != "." {
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	}); err != nil {
		return err
	}
	for i := len(paths) - 1; i >= 0; i-- {
		rel := paths[i]
		if _, ok := source[rel]; ok {
			continue
		}
		target := filepath.Join(localRoot, filepath.FromSlash(rel))
		fmt.Printf("DELETE %s\n", target)
		if !dryRun {
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				_ = os.RemoveAll(target)
			}
		}
	}
	return nil
}
