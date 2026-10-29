package tests

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/cli"
	"github.com/acierto/nexora/internal/runtimeversion"
)

func TestVersionedMCP(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no prebuilt Windows runtime")
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "nexora", Mode: 0700, Size: int64(len(binary))}); err != nil {
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
	sum := sha256.Sum256(archive.Bytes())
	asset := fmt.Sprintf("nexora_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			fmt.Fprint(w, `{"tag_name":"v0.4.0"}`)
		case "/v0.4.0/checksums.txt":
			fmt.Fprintf(w, "%x  %s\n", sum, asset)
		case "/v0.4.0/" + asset:
			_, _ = w.Write(archive.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	if code, out, err := command(t, root, "init"); code != 0 {
		t.Fatal(code, out, err)
	}
	opts := runtimeversion.Options{CacheDir: t.TempDir(), DownloadBase: server.URL, LatestURL: server.URL + "/latest"}
	call := func(selector, config string) (int, string, string) {
		t.Helper()
		var out, diagnostic bytes.Buffer
		input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}` + "\n"
		code := cli.RunWith(context.Background(), []string{"mcp", "--version", selector, "--config", config}, root, cli.Streams{Stdin: strings.NewReader(input), Runtime: opts}, &out, &diagnostic)
		return code, out.String(), diagnostic.String()
	}
	for _, selector := range []string{"v0.4.0", "latest"} {
		code, out, diagnostic := call(selector, filepath.Join(root, "nexora.yaml"))
		var response struct {
			Result struct {
				ServerInfo struct {
					Name string `json:"name"`
				} `json:"serverInfo"`
			} `json:"result"`
		}
		if code != 0 || json.Unmarshal([]byte(out), &response) != nil || response.Result.ServerInfo.Name != "nexora" || !strings.Contains(diagnostic, "using runtime v0.4.0") {
			t.Fatal(code, out, diagnostic)
		}
	}
	code, out, diagnostic := call("v0.4.0", filepath.Join(root, "missing.yaml"))
	if code != 1 || out != "" || !strings.Contains(diagnostic, "Nexora warning") || !strings.Contains(diagnostic, "Continue repository work") {
		t.Fatal("child failure was not preserved", code, out, diagnostic)
	}
	server.Close()
	code, out, diagnostic = call("v0.4.0", filepath.Join(root, "nexora.yaml"))
	if code != 0 || out == "" {
		t.Fatal("offline pin", code, out, diagnostic)
	}
	code, out, diagnostic = call("latest", filepath.Join(root, "nexora.yaml"))
	if code != 1 || out != "" || !strings.Contains(diagnostic, "resolve latest release") {
		t.Fatal("latest fallback", code, out, diagnostic)
	}
}
