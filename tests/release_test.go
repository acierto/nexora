package tests

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Run the real release script for the Windows targets and inspect what it
// publishes: zip layout, installers, and checksum coverage.
func TestReleaseBuildWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX release script")
	}
	if _, err := exec.LookPath("zip"); err != nil {
		t.Skip("zip is required for Windows archives")
	}
	// The script writes dist/ under its own checkout; use a copy so the test
	// never replaces a developer's dist/ directory.
	root := t.TempDir()
	for _, path := range []string{"go.mod", "go.sum", "README.md", "LICENSE", "cmd", "internal", "scripts"} {
		if output, err := exec.Command("cp", "-R", filepath.Join(repository, path), filepath.Join(root, path)).CombinedOutput(); err != nil {
			t.Fatalf("copy %s: %v %s", path, err, output)
		}
	}
	cmd := exec.Command("sh", filepath.Join(root, "scripts", "release-build.sh"), "v0.0.0-test")
	cmd.Env = append(os.Environ(), "NEXORA_TARGETS=linux/amd64 windows/amd64 windows/arm64")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("release build: %v\n%s", err, output)
	}
	dist := filepath.Join(root, "dist")
	sums := map[string]string{}
	file, err := os.Open(filepath.Join(dist, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			t.Fatalf("checksum line %q", scanner.Text())
		}
		sums[fields[1]] = fields[0]
	}
	want := []string{"install.ps1", "install.sh", "nexora_linux_amd64.tar.gz", "nexora_windows_amd64.zip", "nexora_windows_arm64.zip"}
	got := make([]string, 0, len(sums))
	for name := range sums {
		got = append(got, name)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("checksums cover %v, want %v", got, want)
	}
	for name, sum := range sums {
		data, err := os.ReadFile(filepath.Join(dist, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != sum {
			t.Fatalf("checksum for %s: %v", name, err)
		}
	}
	for _, arch := range []string{"amd64", "arm64"} {
		reader, err := zip.OpenReader(filepath.Join(dist, "nexora_windows_"+arch+".zip"))
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, member := range reader.File {
			names = append(names, member.Name)
			if member.Name == "nexora.exe" {
				f, err := member.Open()
				if err != nil {
					t.Fatal(err)
				}
				header := make([]byte, 2)
				_, err = f.Read(header)
				f.Close()
				if err != nil || string(header) != "MZ" {
					t.Fatalf("%s nexora.exe is not a Windows executable", arch)
				}
			}
		}
		reader.Close()
		slices.Sort(names)
		if !slices.Equal(names, []string{"LICENSE", "README.md", "nexora.exe"}) {
			t.Fatalf("%s zip contains %v", arch, names)
		}
	}
	if _, err := os.Stat(filepath.Join(dist, "nexora_windows_amd64.tar.gz")); !os.IsNotExist(err) {
		t.Fatal("Windows target also produced a tar.gz")
	}
}
