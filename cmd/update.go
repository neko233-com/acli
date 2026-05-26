package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	updateRepo     = "neko233-com/unicli"
	updateBinary   = "unicli"
	githubAPI      = "https://api.github.com/repos/" + updateRepo + "/releases"
	updateUserAgent = "unicli/" + updateBinary
)

var (
	updateCheckOnly bool
	updateYes       bool
	updateVersion   string
)

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "only check for updates, do not install")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "install without confirmation")
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "install a specific version (e.g. v1.0.5)")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update unicli to the latest release",
	Long: `Download and install the latest unicli release from GitHub.

Examples:
  unicli update              Check and install if a newer version exists
  unicli update --check      Only show current vs latest version
  unicli update -y           Install latest without confirmation
  unicli update --version v1.0.5   Install a specific release`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runUpdate() error {
	fmt.Printf("Current version: %s\n", currentVersion)

	targetTag, err := resolveTargetVersion()
	if err != nil {
		return err
	}

	fmt.Printf("Latest version:  %s\n", targetTag)

	if updateVersion == "" {
		if normalizeVersionTag(targetTag) == normalizeVersionTag(currentVersion) {
			fmt.Println("You are running the latest version.")
			return nil
		}
		if !versionLess(currentVersion, targetTag) {
			fmt.Println("You are running a newer build than the latest release.")
			return nil
		}
	}

	if updateCheckOnly {
		fmt.Println("A newer version is available. Run: unicli update -y")
		return nil
	}

	if !updateYes {
		fmt.Print("Install update? [y/N]: ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	execPath, err := resolveExecutable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}

	url := releaseAssetURL(targetTag)
	fmt.Printf("Downloading %s...\n", url)

	tmpPath, err := downloadReleaseBinary(url)
	if err != nil {
		return err
	}

	fmt.Printf("Installing to %s...\n", execPath)
	if err := applySelfUpdate(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	fmt.Printf("Updated to %s\n", targetTag)
	return nil
}

func resolveTargetVersion() (string, error) {
	if updateVersion != "" {
		return formatVersionTag(updateVersion), nil
	}
	return fetchLatestReleaseTag()
}

func formatVersionTag(v string) string {
	v = strings.TrimSpace(v)
	for strings.HasPrefix(v, "v") {
		v = v[1:]
	}
	if v == "" {
		return "v0.0.0"
	}
	return "v" + v
}

func normalizeVersionTag(v string) string {
	return formatVersionTag(v)
}

func fetchLatestReleaseTag() (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubAPI+"/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", updateUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse release: %w", err)
	}
	if result.TagName == "" {
		return "", fmt.Errorf("empty tag_name in release")
	}
	return result.TagName, nil
}

func releaseAssetURL(tag string) string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	ver := strings.TrimPrefix(tag, "v")
	return fmt.Sprintf(
		"https://github.com/%s/releases/download/v%s/%s-%s-%s%s",
		updateRepo, ver, updateBinary, runtime.GOOS, runtime.GOARCH, ext,
	)
}

func downloadReleaseBinary(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", updateUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("download %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("unicli-update-%d", os.Getpid()))
	if runtime.GOOS == "windows" {
		tmpPath += ".exe"
	}

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmpPath)
		return "", err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func resolveExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func applySelfUpdate(src, dst string) error {
	if runtime.GOOS == "windows" {
		return applySelfUpdateWindows(src, dst)
	}
	return applySelfUpdateUnix(src, dst)
}

func applySelfUpdateUnix(src, dst string) error {
	mode := os.FileMode(0755)
	if info, err := os.Stat(dst); err == nil {
		mode = info.Mode()
	}
	if err := os.Chmod(src, mode); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		return fmt.Errorf("replace binary (try with sudo): %w", err)
	}
	_ = os.Remove(src)
	return nil
}

func applySelfUpdateWindows(src, dst string) error {
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("unicli-update-%d.cmd", os.Getpid()))
	script := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move /Y "%s" "%s" >nul
if errorlevel 1 exit /b 1
start "" "%s"
del "%%~f0"
`, src, dst, dst)

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return err
	}

	c := exec.Command("cmd", "/c", "start", "", "/min", scriptPath)
	if err := c.Start(); err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("start updater: %w", err)
	}

	fmt.Println("Update will finish after this process exits.")
	os.Exit(0)
	return nil
}

// versionLess returns true if a < b (semver-ish, numeric segments only).
func versionLess(a, b string) bool {
	pa := strings.Split(normalizeVersionTag(a), ".")
	pb := strings.Split(normalizeVersionTag(b), ".")
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		ai, bi := 0, 0
		if i < len(pa) {
			ai, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			bi, _ = strconv.Atoi(pb[i])
		}
		if ai != bi {
			return ai < bi
		}
	}
	return false
}
