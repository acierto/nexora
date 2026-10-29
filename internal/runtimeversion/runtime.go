// Package runtimeversion selects verified release executables for MCP clients.
package runtimeversion

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var releaseTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-.][0-9A-Za-z.-]+)?$`)

// Options provides the filesystem, platform and HTTP boundaries used by tests.
// Empty values select the public release service and the current user's cache.
type Options struct {
	CacheDir, DownloadBase, LatestURL, OS, Arch string
	HTTP                                        *http.Client
}

type Executable struct{ Path, Version string }
type manifest struct{ Version, SHA256 string }

func ValidSelector(selector string) bool {
	return selector == "latest" || releaseTag.MatchString(selector)
}

func Resolve(ctx context.Context, selector string, opts Options) (Executable, error) {
	if !ValidSelector(selector) {
		return Executable{}, fmt.Errorf("version must be latest or an exact release tag such as v0.4.0")
	}
	if opts.HTTP == nil {
		opts.HTTP = &http.Client{Timeout: 60 * time.Second}
	}
	if opts.OS == "" {
		opts.OS = runtime.GOOS
	}
	if opts.Arch == "" {
		opts.Arch = runtime.GOARCH
	}
	if (opts.OS != "darwin" && opts.OS != "linux") || (opts.Arch != "arm64" && opts.Arch != "amd64") {
		return Executable{}, fmt.Errorf("no prebuilt Nexora runtime for %s/%s", opts.OS, opts.Arch)
	}
	if opts.DownloadBase == "" {
		opts.DownloadBase = "https://github.com/acierto/nexora/releases/download"
	}
	if opts.LatestURL == "" {
		opts.LatestURL = "https://api.github.com/repos/acierto/nexora/releases/latest"
	}
	tag := selector
	if selector == "latest" {
		data, err := download(ctx, opts.HTTP, opts.LatestURL, 1<<20)
		if err != nil {
			return Executable{}, fmt.Errorf("resolve latest release: %w", err)
		}
		var release struct {
			Tag string `json:"tag_name"`
		}
		if err := json.Unmarshal(data, &release); err != nil {
			return Executable{}, fmt.Errorf("resolve latest release: %w", err)
		}
		if !releaseTag.MatchString(release.Tag) {
			return Executable{}, fmt.Errorf("latest release returned an invalid tag")
		}
		tag = release.Tag
	}
	if opts.CacheDir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return Executable{}, err
		}
		opts.CacheDir = filepath.Join(base, "nexora", "runtimes")
	}
	dir := filepath.Join(opts.CacheDir, tag, opts.OS+"_"+opts.Arch)
	if _, err := os.Lstat(dir); err == nil {
		return cached(dir, tag)
	} else if !os.IsNotExist(err) {
		return Executable{}, err
	}
	asset := fmt.Sprintf("nexora_%s_%s.tar.gz", opts.OS, opts.Arch)
	base := strings.TrimRight(opts.DownloadBase, "/") + "/" + tag + "/"
	checksums, err := download(ctx, opts.HTTP, base+"checksums.txt", 1<<20)
	if err != nil {
		return Executable{}, fmt.Errorf("release %s checksums: %w", tag, err)
	}
	expected, err := checksum(checksums, asset)
	if err != nil {
		return Executable{}, err
	}
	archive, err := download(ctx, opts.HTTP, base+asset, 128<<20)
	if err != nil {
		return Executable{}, fmt.Errorf("release %s archive: %w", tag, err)
	}
	if digest(archive) != expected {
		return Executable{}, fmt.Errorf("release %s archive checksum mismatch", tag)
	}
	executable, err := extract(archive)
	if err != nil {
		return Executable{}, err
	}
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return Executable{}, err
	}
	temporary, err := os.MkdirTemp(parent, ".install-")
	if err != nil {
		return Executable{}, err
	}
	defer os.RemoveAll(temporary)
	if err := os.WriteFile(filepath.Join(temporary, "nexora"), executable, 0700); err != nil {
		return Executable{}, err
	}
	metadata, _ := json.Marshal(manifest{Version: tag, SHA256: digest(executable)})
	if err := os.WriteFile(filepath.Join(temporary, "manifest.json"), metadata, 0600); err != nil {
		return Executable{}, err
	}
	if err := os.Rename(temporary, dir); err != nil {
		// Another installer may have completed the same immutable version first.
		if winner, checkErr := cached(dir, tag); checkErr == nil {
			return winner, nil
		}
		return Executable{}, err
	}
	return cached(dir, tag)
}

func cached(dir, tag string) (Executable, error) {
	fail := func() (Executable, error) {
		return Executable{}, fmt.Errorf("runtime cache integrity check failed for %s; remove that version's cache directory and retry", dir)
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fail()
	}
	for _, name := range []string{"nexora", "manifest.json"} {
		info, err := os.Lstat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() || info.Size() > 128<<20 || (name == "manifest.json" && info.Size() > 4096) {
			return fail()
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return fail()
	}
	var metadata manifest
	if json.Unmarshal(data, &metadata) != nil || metadata.Version != tag {
		return fail()
	}
	executable, err := os.ReadFile(filepath.Join(dir, "nexora"))
	if err != nil || digest(executable) != metadata.SHA256 {
		return fail()
	}
	return Executable{Path: filepath.Join(dir, "nexora"), Version: tag}, nil
}
func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func checksum(data []byte, asset string) (string, error) {
	expected := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != asset {
			continue
		}
		decoded, err := hex.DecodeString(fields[0])
		if expected != "" || err != nil || len(decoded) != sha256.Size {
			return "", fmt.Errorf("invalid checksum entry for %s", asset)
		}
		expected = strings.ToLower(fields[0])
	}
	if expected == "" {
		return "", fmt.Errorf("release checksums have no entry for %s", asset)
	}
	return expected, nil
}
func download(ctx context.Context, client *http.Client, url string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("release response exceeds %d bytes", limit)
	}
	return data, nil
}
func extract(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	var executable []byte
	for {
		member, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if member.Name != "nexora" {
			continue
		}
		if (member.Typeflag != tar.TypeReg && member.Typeflag != tar.TypeRegA) || member.Size <= 0 || member.Size > 128<<20 || executable != nil {
			return nil, fmt.Errorf("invalid Nexora executable in release archive")
		}
		executable, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
	}
	if executable == nil {
		return nil, fmt.Errorf("Nexora executable missing from release archive")
	}
	return executable, nil
}
