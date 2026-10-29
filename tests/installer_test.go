package tests

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Use the real CLI built by TestMain in a real release archive. No GitHub or
// credentials are needed to exercise download, verification and installation.
func TestInstaller(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX installer")
	}
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl certificate diagnostics")
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "nexora", Mode: 0755, Size: int64(len(binary))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(binary); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	asset := fmt.Sprintf("nexora_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, tc := range []struct {
		name                string
		tls, trust, corrupt bool
	}{
		{name: "verified archive"},
		{name: "checksum failure", corrupt: true},
		{name: "untrusted certificate", tls: true},
		{name: "configured certificate trust", tls: true, trust: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checksum := fmt.Sprintf("%x", sha256.Sum256(archive.Bytes()))
			if tc.corrupt {
				checksum = strings.Repeat("0", 64)
			}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/" + asset:
					_, _ = w.Write(archive.Bytes())
				case "/checksums.txt":
					_, _ = fmt.Fprintf(w, "%s  %s\n", checksum, asset)
				default:
					http.NotFound(w, r)
				}
			})
			var server *httptest.Server
			if tc.tls {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			defer server.Close()
			dir := t.TempDir()
			target := filepath.Join(dir, "install with spaces")
			cmd := exec.Command("sh", filepath.Join(repository, "scripts", "install.sh"))
			// Do not inherit a caller's CA override; each test controls trust.
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "CURL_CA_BUNDLE=") && !strings.HasPrefix(entry, "SSL_CERT_FILE=") && !strings.HasPrefix(entry, "SSL_CERT_DIR=") && !strings.HasPrefix(entry, "NEXORA_") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "NEXORA_VERSION=v0.3.1", "NEXORA_INSTALL_DIR="+target, "NEXORA_DOWNLOAD_BASE="+server.URL, "NO_PROXY=127.0.0.1,localhost", "no_proxy=127.0.0.1,localhost")
			if tc.trust {
				ca := filepath.Join(dir, "ca.pem")
				if err := os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
					t.Fatal(err)
				}
				cmd.Env = append(cmd.Env, "CURL_CA_BUNDLE="+ca)
			}
			output, err := cmd.CombinedOutput()
			failure := tc.corrupt || tc.tls && !tc.trust
			installed := filepath.Join(target, "nexora")
			if failure {
				if err == nil {
					t.Fatalf("expected failure: %s", output)
				}
				if _, err := os.Stat(installed); !os.IsNotExist(err) {
					t.Fatalf("installed after failed verification: %v", err)
				}
				expected := "checksum mismatch"
				if tc.tls {
					expected = "HTTPS certificate trust failed"
				}
				if !strings.Contains(string(output), expected) {
					t.Fatalf("missing diagnostic %q: %s", expected, output)
				}
			} else {
				if err != nil {
					t.Fatalf("installer: %v\n%s", err, output)
				}
				result, err := exec.Command(installed, "version").CombinedOutput()
				if err != nil || !strings.Contains(string(result), "nexora") {
					t.Fatalf("installed CLI: %v %s", err, result)
				}
			}
		})
	}
}
