package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type downloadProbe struct {
	URL       string `json:"url"`
	LatencyMS int64  `json:"latency_ms"`
	Size      int64  `json:"size"`
	Range     bool   `json:"range"`
	Error     string `json:"error,omitempty"`
}

var downloadCmd = &cobra.Command{Use: "download", Short: "Mirror-aware, multi-thread local downloader"}
var downloadLatencyCmd = &cobra.Command{Use: "latency <url...>", Short: "Probe download mirrors and return fastest first", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	probes := probeDownloads(cmd.Context(), args)
	return json.NewEncoder(os.Stdout).Encode(probes)
}}
var downloadGetCmd = &cobra.Command{Use: "get <url> <out>", Short: "Probe mirrors, then download with CPU×2 threads", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	mirrors, _ := cmd.Flags().GetStringSlice("mirror")
	urls := append([]string{args[0]}, mirrors...)
	probes := probeDownloads(cmd.Context(), urls)
	sort.SliceStable(probes, func(i, j int) bool { return probes[i].LatencyMS < probes[j].LatencyMS })
	var selected *downloadProbe
	for i := range probes {
		if probes[i].Error == "" {
			selected = &probes[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("all download mirrors failed: %v", probes)
	}
	threads, _ := cmd.Flags().GetInt("threads")
	if threads <= 0 {
		threads = runtime.NumCPU() * 2
	}
	sum, _ := cmd.Flags().GetString("sha256")
	return downloadHTTPFile(cmd.Context(), *selected, args[1], threads, sum)
}}

func probeDownloads(ctx context.Context, urls []string) []downloadProbe {
	out := make([]downloadProbe, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) { defer wg.Done(); out[i] = probeDownload(ctx, url) }(i, url)
	}
	wg.Wait()
	return out
}
func probeDownload(ctx context.Context, url string) downloadProbe {
	result := downloadProbe{URL: url}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	request.Header.Set("Range", "bytes=0-0")
	started := time.Now()
	response, err := http.DefaultClient.Do(request)
	result.LatencyMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		result.Error = response.Status
		return result
	}
	result.Range = response.StatusCode == http.StatusPartialContent
	if response.ContentLength > 0 {
		result.Size = response.ContentLength
	}
	if contentRange := response.Header.Get("Content-Range"); contentRange != "" {
		var end, total int64
		if _, err := fmt.Sscanf(contentRange, "bytes 0-%d/%d", &end, &total); err == nil {
			result.Size = total
			result.Range = true
		}
	}
	return result
}
func downloadHTTPFile(ctx context.Context, probe downloadProbe, output string, threads int, expected string) error {
	if probe.Size <= 0 || !probe.Range {
		threads = 1
	}
	if threads < 1 {
		threads = 1
	}
	if threads > 64 {
		threads = 64
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(output), ".acli-download-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if probe.Size > 0 {
		if err := temp.Truncate(probe.Size); err != nil {
			temp.Close()
			return err
		}
	}
	if threads == 1 {
		err = downloadRange(ctx, probe.URL, temp, 0, probe.Size-1)
	} else {
		err = downloadParallel(ctx, probe.URL, temp, probe.Size, threads)
	}
	if closeErr := temp.Close(); err != nil {
		return err
	} else if closeErr != nil {
		return closeErr
	}
	if expected != "" {
		if err := checkSHA256(tempName, expected); err != nil {
			return err
		}
	}
	return os.Rename(tempName, output)
}
func downloadParallel(ctx context.Context, url string, file *os.File, size int64, threads int) error {
	chunk := (size + int64(threads) - 1) / int64(threads)
	var wg sync.WaitGroup
	errors := make(chan error, threads)
	for start := int64(0); start < size; start += chunk {
		end := start + chunk - 1
		if end >= size {
			end = size - 1
		}
		wg.Add(1)
		go func(start, end int64) { defer wg.Done(); errors <- downloadRange(ctx, url, file, start, end) }(start, end)
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
func downloadRange(ctx context.Context, url string, file *os.File, start, end int64) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if end >= start {
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download %s: %s", url, response.Status)
	}
	writer := io.NewOffsetWriter(file, start)
	_, err = io.Copy(writer, response.Body)
	return err
}
func checkSHA256(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	expected = strings.ToLower(strings.TrimSpace(expected))
	if actual != expected {
		return fmt.Errorf("sha256 mismatch: got %s", actual)
	}
	return nil
}
func init() {
	downloadGetCmd.Flags().StringSlice("mirror", nil, "mirror URL; repeatable, fastest source wins")
	downloadGetCmd.Flags().Int("threads", 0, "parallel ranges; default CPU×2, max 64")
	downloadGetCmd.Flags().String("sha256", "", "expected SHA-256")
	downloadCmd.AddCommand(downloadLatencyCmd, downloadGetCmd)
}
