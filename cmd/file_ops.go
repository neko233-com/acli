package cmd

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{Use: "file", Short: "Cross-platform file CRUD with explicit destructive actions"}

var fileListCmd = &cobra.Command{Use: "list [path]", Short: "List directory entries", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) == 1 {
		path = args[0]
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return json.NewEncoder(os.Stdout).Encode(names)
}}
var fileReadCmd = &cobra.Command{Use: "read <file>", Short: "Write file bytes to stdout", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(os.Stdout, file)
	return err
}}
var fileWriteCmd = &cobra.Command{Use: "write <file> <text...>", Short: "Create text file; replacement requires --force", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	if _, err := os.Stat(args[0]); err == nil && !force {
		return fmt.Errorf("file exists; pass --force to replace")
	}
	if err := os.MkdirAll(filepath.Dir(args[0]), 0o755); err != nil {
		return err
	}
	return os.WriteFile(args[0], []byte(strings.Join(args[1:], " ")+"\n"), 0o644)
}}
var fileAppendCmd = &cobra.Command{Use: "append <file> <text...>", Short: "Append text file", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	if err := os.MkdirAll(filepath.Dir(args[0]), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(args[0], os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintln(file, strings.Join(args[1:], " "))
	return err
}}
var fileCopyCmd = &cobra.Command{Use: "copy <source> <destination>", Short: "Copy file", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return copyLocalFile(args[0], args[1]) }}
var fileMoveCmd = &cobra.Command{Use: "move <source> <destination>", Short: "Move file or directory", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	if err := os.MkdirAll(filepath.Dir(args[1]), 0o755); err != nil {
		return err
	}
	return os.Rename(args[0], args[1])
}}
var fileRemoveCmd = &cobra.Command{Use: "remove <path>", Short: "Permanently remove file/directory; requires --yes", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		return fmt.Errorf("permanent deletion requires --yes")
	}
	target, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if target == filepath.VolumeName(target)+string(os.PathSeparator) || target == cwd {
		return fmt.Errorf("refusing filesystem root or current directory")
	}
	return os.RemoveAll(target)
}}
var fileHashCmd = &cobra.Command{Use: "hash <file>", Short: "SHA-256 file hash", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	fmt.Printf("%x  %s\n", hash.Sum(nil), args[0])
	return nil
}}
var filePatchCmd = &cobra.Command{Use: "patch <file> <old> <new>", Short: "Exact text replace; requires --out or --yes", Args: cobra.ExactArgs(3), RunE: func(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), args[1]) {
		return fmt.Errorf("text not found")
	}
	updated := []byte(strings.ReplaceAll(string(data), args[1], args[2]))
	out, _ := cmd.Flags().GetString("out")
	yes, _ := cmd.Flags().GetBool("yes")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"file": args[0], "changed": true, "bytes_before": len(data), "bytes_after": len(updated)})
	}
	if out == "" {
		if !yes {
			return fmt.Errorf("choose --out <file> or --yes for in-place patch")
		}
		out = args[0]
	}
	return os.WriteFile(out, updated, 0o644)
}}
var fileArchiveCmd = &cobra.Command{Use: "archive <out.zip> <path...>", Short: "Create ZIP archive", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return createZip(args[0], args[1:]) }}
var fileUnarchiveCmd = &cobra.Command{Use: "unarchive <archive.zip> <out-dir>", Short: "Extract ZIP archive with Zip Slip protection", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return extractZip(args[0], args[1]) }}

func copyLocalFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("source is a directory; use acli sync for directory trees")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func createZip(output string, inputs []string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()
	for _, input := range inputs {
		base := filepath.Base(input)
		if err := filepath.Walk(input, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(input, path)
			entry, err := writer.Create(filepath.ToSlash(filepath.Join(base, rel)))
			if err != nil {
				return err
			}
			source, err := os.Open(path)
			if err != nil {
				return err
			}
			defer source.Close()
			_, err = io.Copy(entry, source)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}
func extractZip(input, output string) error {
	reader, err := zip.OpenReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()
	root, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	for _, entry := range reader.File {
		target := filepath.Join(root, entry.Name)
		if !strings.HasPrefix(filepath.Clean(target), root+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe archive path %q", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := entry.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entry.Mode())
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		closeErr := dst.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func init() {
	fileWriteCmd.Flags().Bool("force", false, "replace existing file")
	fileRemoveCmd.Flags().Bool("yes", false, "confirm permanent removal")
	fileCmd.AddCommand(fileListCmd, fileReadCmd, fileWriteCmd, fileAppendCmd, fileCopyCmd, fileMoveCmd, fileRemoveCmd, fileHashCmd, filePatchCmd, fileArchiveCmd, fileUnarchiveCmd)
}
