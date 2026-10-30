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
	"sync/atomic"
	"testing"
)

// Use a real versioned CLI in a real release archive. No GitHub or
// credentials are needed to exercise download, verification and installation.
func TestInstaller(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX installer")
	}
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl certificate diagnostics")
	}
	releaseBinary := func(version string) []byte {
		t.Helper()
		path := filepath.Join(t.TempDir(), "nexora")
		cmd := exec.Command("go", "build", "-ldflags", "-X github.com/acierto/nexora/internal/cli.Version="+version, "-o", path, "./cmd/nexora")
		cmd.Dir = repository
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build: %v %s", err, output)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	versions := map[string][]byte{}
	for _, version := range []string{"v0.4.0", "v0.4.1", "v0.10.0", "v1.0.0", "v0.4.1-rc.1"} {
		versions[version] = releaseBinary(version)
	}
	dev, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	versions["dev"] = dev
	binary := versions["v0.4.1"]
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
		name                               string
		tls, trust, corrupt                bool
		ensure, reuse, mismatch, directory bool
		existing                           string
	}{
		{name: "verified archive"},
		{name: "checksum failure", corrupt: true},
		{name: "untrusted certificate", tls: true},
		{name: "configured certificate trust", tls: true, trust: true},
		{name: "ensure fresh", ensure: true},
		{name: "ensure upgrade", ensure: true, existing: "v0.4.0"},
		{name: "ensure current offline", ensure: true, existing: "v0.4.1", reuse: true},
		{name: "ensure newer minor offline", ensure: true, existing: "v0.10.0", reuse: true},
		{name: "ensure newer major offline", ensure: true, existing: "v1.0.0", reuse: true},
		{name: "ensure prerelease", ensure: true, existing: "v0.4.1-rc.1"},
		{name: "ensure development", ensure: true, existing: "dev"},
		{name: "explicit downgrade", existing: "v1.0.0"},
		{name: "failed upgrade preserves installation", ensure: true, existing: "v0.4.0", corrupt: true},
		{name: "wrong executable version preserves installation", existing: "v0.4.0", mismatch: true},
		{name: "directory target rejected", directory: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checksum := fmt.Sprintf("%x", sha256.Sum256(archive.Bytes()))
			if tc.corrupt {
				checksum = strings.Repeat("0", 64)
			}
			var requests atomic.Int32
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
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
			installed := filepath.Join(target, "nexora")
			if tc.existing != "" || tc.directory {
				if err := os.MkdirAll(target, 0755); err != nil {
					t.Fatal(err)
				}
				if tc.directory {
					if err := os.Mkdir(installed, 0755); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(installed, versions[tc.existing], 0755); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{filepath.Join(repository, "scripts", "install.sh")}
			if tc.ensure {
				args = append(args, "--ensure")
			}
			cmd := exec.Command("sh", args...)
			// Do not inherit a caller's CA override; each test controls trust.
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "CURL_CA_BUNDLE=") && !strings.HasPrefix(entry, "SSL_CERT_FILE=") && !strings.HasPrefix(entry, "SSL_CERT_DIR=") && !strings.HasPrefix(entry, "NEXORA_") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			pin := "v0.4.1"
			if tc.mismatch {
				pin = "v0.4.2"
			}
			cmd.Env = append(cmd.Env, "NEXORA_VERSION="+pin, "NEXORA_INSTALL_DIR="+target, "NEXORA_DOWNLOAD_BASE="+server.URL, "NO_PROXY=127.0.0.1,localhost", "no_proxy=127.0.0.1,localhost")
			if tc.trust {
				ca := filepath.Join(dir, "ca.pem")
				if err := os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
					t.Fatal(err)
				}
				cmd.Env = append(cmd.Env, "CURL_CA_BUNDLE="+ca)
			}
			output, err := cmd.CombinedOutput()
			failure := tc.corrupt || tc.tls && !tc.trust || tc.mismatch || tc.directory
			if failure {
				if err == nil {
					t.Fatalf("expected failure: %s", output)
				}
				if tc.existing != "" {
					data, err := os.ReadFile(installed)
					if err != nil || !bytes.Equal(data, versions[tc.existing]) {
						t.Fatalf("previous executable changed: %v", err)
					}
				} else if !tc.directory {
					if _, err := os.Stat(installed); !os.IsNotExist(err) {
						t.Fatalf("installed after failed verification: %v", err)
					}
				}
				expected := "checksum mismatch"
				if tc.mismatch {
					expected = "expected nexora v0.4.2"
				}
				if tc.directory {
					expected = "destination is a directory"
				}
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
				want := "nexora v0.4.1\n"
				if tc.reuse {
					want = "nexora " + tc.existing + "\n"
				}
				if err != nil || string(result) != want {
					t.Fatalf("installed CLI: %v %s", err, result)
				}
			}
			if tc.reuse && requests.Load() != 0 {
				t.Fatalf("reuse performed %d network requests", requests.Load())
			}
			leftovers, err := filepath.Glob(filepath.Join(target, ".nexora-install.*"))
			if err != nil || len(leftovers) != 0 {
				t.Fatalf("staging not cleaned: %v %v", leftovers, err)
			}
		})
	}
	for _, pin := range []string{"latest", "dev", "v0.4.1-rc.1", "v999999999.0.0"} {
		t.Run("ensure rejects "+pin, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "absent")
			cmd := exec.Command("sh", filepath.Join(repository, "scripts", "install.sh"), "--ensure")
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "NEXORA_VERSION=" + pin, "NEXORA_INSTALL_DIR=" + target}
			output, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "requires a stable NEXORA_VERSION") {
				t.Fatalf("invalid pin: %v %s", err, output)
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatalf("invalid pin modified target: %v", err)
			}
		})
	}
}
