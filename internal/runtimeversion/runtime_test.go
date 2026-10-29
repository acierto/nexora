package runtimeversion

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func archive(t *testing.T, name string, kind byte, payload []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	header := &tar.Header{Name: name, Typeflag: kind, Mode: 0700}
	if kind == tar.TypeReg {
		header.Size = int64(len(payload))
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if kind == tar.TypeReg {
		if _, err := tw.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestVerifiedRuntimeSelection(t *testing.T) {
	payload := []byte("release executable bytes")
	data := archive(t, "nexora", tar.TypeReg, payload)
	var lock sync.Mutex
	tag := "v0.4.0"
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		requests = append(requests, r.URL.Path)
		switch {
		case r.URL.Path == "/latest":
			fmt.Fprintf(w, `{"tag_name":%q}`, tag)
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			fmt.Fprintf(w, "%s  nexora_linux_amd64.tar.gz\n", digest(data))
		case strings.HasSuffix(r.URL.Path, "nexora_linux_amd64.tar.gz"):
			_, _ = w.Write(data)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	opts := Options{CacheDir: t.TempDir(), DownloadBase: server.URL, LatestURL: server.URL + "/latest", OS: "linux", Arch: "amd64"}
	first, err := Resolve(context.Background(), "v0.4.0", opts)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(first.Path)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatal(string(got), err)
	}
	lock.Lock()
	if len(requests) != 2 {
		t.Fatal(requests)
	}
	requests = nil
	lock.Unlock()
	second, err := Resolve(context.Background(), "v0.4.0", opts)
	if err != nil || first != second {
		t.Fatal(second, err)
	}
	lock.Lock()
	if len(requests) != 0 {
		t.Fatal("cached pin made network requests", requests)
	}
	lock.Unlock()
	latest, err := Resolve(context.Background(), "latest", opts)
	if err != nil || latest != first {
		t.Fatal(latest, err)
	}
	lock.Lock()
	tag = "v0.4.1"
	lock.Unlock()
	latest, err = Resolve(context.Background(), "latest", opts)
	if err != nil || latest.Version != "v0.4.1" || latest.Path == first.Path {
		t.Fatal(latest, err)
	}
	server.Close()
	if _, err := Resolve(context.Background(), "latest", opts); err == nil {
		t.Fatal("latest silently fell back after network failure")
	}
	if selected, err := Resolve(context.Background(), "v0.4.0", opts); err != nil || selected != first {
		t.Fatal("offline pin", selected, err)
	}
	if err := os.WriteFile(first.Path, []byte("corrupted"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(context.Background(), "v0.4.0", opts); err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatal(err)
	}
}

func TestRuntimeRefusals(t *testing.T) {
	valid := archive(t, "nexora", tar.TypeReg, []byte("binary"))
	for _, tc := range []struct {
		name, selector, os, checksum, latest string
		data                                 []byte
	}{
		{name: "checksum mismatch", selector: "v0.4.0", data: valid, checksum: strings.Repeat("0", 64)},
		{name: "missing checksum", selector: "v0.4.0", data: valid, checksum: "missing"},
		{name: "missing executable", selector: "v0.4.0", data: archive(t, "../nexora", tar.TypeReg, []byte("binary"))},
		{name: "symlink executable", selector: "v0.4.0", data: archive(t, "nexora", tar.TypeSymlink, nil)},
		{name: "invalid latest tag", selector: "latest", data: valid, latest: `{"tag_name":"../../escape"}`},
		{name: "invalid selector", selector: "../escape", data: valid},
		{name: "empty selector", selector: "", data: valid},
		{name: "unsupported platform", selector: "v0.4.0", os: "windows", data: valid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path == "/latest" {
					fmt.Fprint(w, tc.latest)
					return
				}
				if strings.HasSuffix(r.URL.Path, "checksums.txt") {
					checksum := tc.checksum
					if checksum == "" {
						checksum = digest(tc.data)
					}
					fmt.Fprintf(w, "%s  nexora_linux_amd64.tar.gz\n", checksum)
					return
				}
				_, _ = w.Write(tc.data)
			}))
			defer server.Close()
			platform := tc.os
			if platform == "" {
				platform = "linux"
			}
			cache := t.TempDir()
			_, err := Resolve(context.Background(), tc.selector, Options{CacheDir: cache, DownloadBase: server.URL, LatestURL: server.URL + "/latest", OS: platform, Arch: "amd64"})
			if err == nil {
				t.Fatal("expected refusal")
			}
			if (!ValidSelector(tc.selector) || platform == "windows") && calls.Load() != 0 {
				t.Fatal("refusal after network request")
			}
			if _, err := os.Stat(filepath.Join(cache, "v0.4.0", "linux_amd64", "nexora")); !os.IsNotExist(err) {
				t.Fatal("installed a refused artifact", err)
			}
		})
	}
}

func TestConcurrentRuntimeInstall(t *testing.T) {
	data := archive(t, "nexora", tar.TypeReg, []byte("binary"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") {
			fmt.Fprintf(w, "%s  nexora_linux_amd64.tar.gz\n", digest(data))
			return
		}
		_, _ = w.Write(data)
	}))
	defer server.Close()
	opts := Options{CacheDir: t.TempDir(), DownloadBase: server.URL, OS: "linux", Arch: "amd64"}
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Resolve(context.Background(), "v0.4.0", opts); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
