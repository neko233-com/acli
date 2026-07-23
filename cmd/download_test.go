package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func rangeServer(body []byte, delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		start, end := 0, len(body)-1
		if value := request.Header.Get("Range"); value != "" {
			_, _ = fmt.Sscanf(value, "bytes=%d-%d", &start, &end)
			if end < 0 || end >= len(body) {
				end = len(body) - 1
			}
			writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(body)))
			writer.WriteHeader(http.StatusPartialContent)
		}
		writer.Header().Set("Content-Length", strconv.Itoa(end-start+1))
		_, _ = writer.Write(body[start : end+1])
	}))
}

func TestProbeChoosesFastMirrorAndParallelDownload(t *testing.T) {
	body := []byte(strings.Repeat("acli-download-data-", 4096))
	slow, fast := rangeServer(body, 30*time.Millisecond), rangeServer(body, 0)
	defer slow.Close()
	defer fast.Close()
	probes := probeDownloads(context.Background(), []string{slow.URL, fast.URL})
	if probes[0].Error != "" || probes[1].Error != "" || !probes[1].Range || probes[1].Size != int64(len(body)) {
		t.Fatalf("bad probes: %#v", probes)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(body))
	output := filepath.Join(t.TempDir(), "download.bin")
	if err := downloadHTTPFile(context.Background(), probes[1], output, 8, hash); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("download mismatch: %d != %d", len(got), len(body))
	}
}

func TestDownloadChecksumMismatchKeepsDestinationAbsent(t *testing.T) {
	server := rangeServer([]byte("ok"), 0)
	defer server.Close()
	probe := probeDownload(context.Background(), server.URL)
	output := filepath.Join(t.TempDir(), "bad.bin")
	if err := downloadHTTPFile(context.Background(), probe, output, 2, strings.Repeat("0", 64)); err == nil {
		t.Fatal("expected checksum mismatch")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}

func TestDownloadWithoutRangeFallsBackToSingleStream(t *testing.T) {
	body := []byte("single stream content")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = writer.Write(body)
	}))
	defer server.Close()
	probe := probeDownload(context.Background(), server.URL)
	if probe.Error != "" || probe.Range || probe.Size != int64(len(body)) {
		t.Fatalf("unexpected probe: %#v", probe)
	}
	output := filepath.Join(t.TempDir(), "single.bin")
	if err := downloadHTTPFile(context.Background(), probe, output, 32, ""); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil || string(got) != string(body) {
		t.Fatalf("body=%q err=%v", got, err)
	}
}

func TestDownloadProbeAndRangeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	if probe := probeDownload(context.Background(), server.URL); probe.Error != "503 Service Unavailable" {
		t.Fatalf("probe=%#v", probe)
	}
	if err := downloadRange(context.Background(), server.URL, os.Stdout, 0, 1); err == nil {
		t.Fatal("expected failed range request")
	}
}
