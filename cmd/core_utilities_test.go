package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPortParsers(t *testing.T) {
	windows := parseWindowsNetstat("  TCP    127.0.0.1:80  0.0.0.0:0  LISTENING  123\n  UDP    0.0.0.0:53  *:*  456\n")
	if len(windows) != 2 || windows[0].PID != "123" || windows[1].Protocol != "udp" || windows[1].PID != "456" {
		t.Fatalf("windows parse: %#v", windows)
	}
	linux := parseSSPorts("Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process\ntcp LISTEN 0 4096 127.0.0.1:80 0.0.0.0:* users:((\"acli\",pid=42,fd=3))\n")
	if len(linux) != 1 || linux[0].Protocol != "tcp" || linux[0].Local != "127.0.0.1:80" || linux[0].PID != "42" {
		t.Fatalf("ss parse: %#v", linux)
	}
	mac := parseLsofPorts("COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME\nnode 777 me 1u IPv4 0x0 0t0 TCP *:3000 (LISTEN)\n")
	if len(mac) != 1 || mac[0].Process != "node" || mac[0].PID != "777" || mac[0].Local != "*:3000" {
		t.Fatalf("lsof parse: %#v", mac)
	}
}

func TestProcessParsersAndTree(t *testing.T) {
	pipes := parsePipeProcesses("1|0|init|1024|init --serve\n2|1|worker|2048|worker --job\ninvalid")
	if len(pipes) != 2 || pipes[1].Command != "worker --job" {
		t.Fatalf("pipe parse: %#v", pipes)
	}
	ps := parsePSProcesses("1 0 0.1 0.2 init init --serve\n2 1 1.1 2.2 worker worker --job")
	if len(ps) != 2 || ps[0].CPU != "0.1" || ps[1].Name != "worker" {
		t.Fatalf("ps parse: %#v", ps)
	}
	tree, ok := findProcessTree(ps, 1)
	if !ok || len(tree.Children) != 1 || tree.Children[0].PID != 2 {
		t.Fatalf("tree: %#v ok=%v", tree, ok)
	}
	if descendants := processDescendants(ps, 1); len(descendants) != 1 || descendants[0].PID != 2 {
		t.Fatalf("descendants: %#v", descendants)
	}
	if _, ok := findProcessTree(ps, 99); ok {
		t.Fatal("missing process unexpectedly found")
	}
}

func TestSSHHelpersAndGeneratedKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys", "id_ed25519")
	if err := generateEd25519Key(path); err != nil {
		t.Fatal(err)
	}
	if _, err := signerFromFile(path); err != nil {
		t.Fatalf("generated key unreadable: %v", err)
	}
	public, err := publicKeyFromPrivateKeyFile(path)
	if err != nil || !strings.HasPrefix(string(public), "ssh-ed25519 ") {
		t.Fatalf("public key=%q err=%v", public, err)
	}
	if got := shellQuote("it's safe"); got != "'it'\\''s safe'" {
		t.Fatalf("quote=%q", got)
	}
	if got := expandHome("relative/path"); got != "relative/path" {
		t.Fatalf("expand changed relative path: %q", got)
	}
	if got := expandHome("~"); got == "~" || got == "" {
		t.Fatalf("home not expanded: %q", got)
	}
	if !profileInGroup(sshProfile{Groups: []string{"prod", "edge"}}, "edge") {
		t.Fatal("group not found")
	}
	var output bytes.Buffer
	writer := &prefixWriter{prefix: "host: ", writer: &output}
	if _, err := writer.Write([]byte("one\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("two")); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "host: one\nhost: two" {
		t.Fatalf("prefix output=%q", got)
	}
}

func TestSyncPathAndComparisonHelpers(t *testing.T) {
	if got := joinRemote("/srv/", "/app/file.txt"); got != "/srv/app/file.txt" {
		t.Fatalf("join=%q", got)
	}
	if got := pathDir("/srv/app/file.txt"); got != "/srv/app" {
		t.Fatalf("dir=%q", got)
	}
	if got := pathDir("file.txt"); got != "." {
		t.Fatalf("bare dir=%q", got)
	}
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	if err := os.WriteFile(a, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(a, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(b, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	aInfo, _ := os.Stat(a)
	bInfo, _ := os.Stat(b)
	if !sameFile(aInfo, bInfo) {
		t.Fatal("matching files not equal")
	}
	if err := os.WriteFile(b, []byte("different"), 0o644); err != nil {
		t.Fatal(err)
	}
	bInfo, _ = os.Stat(b)
	if sameFile(aInfo, bInfo) {
		t.Fatal("different files equal")
	}
	closer := &sshClientCloser{close: func() error { return nil }}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := (*sshClientCloser)(nil).Close(); err != nil {
		t.Fatal(err)
	}
}
