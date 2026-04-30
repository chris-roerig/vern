package tests

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
)

// #1: Version validation rejects path traversal
func TestVersionValidation(t *testing.T) {
	valid := []string{"1.0.0", "3.13.13", "0.14", "21", "1.95.0"}
	for _, v := range valid {
		if !config.IsValidVersion(v) {
			t.Errorf("IsValidVersion(%q) = false, want true", v)
		}
	}

	invalid := []string{
		"../../etc/passwd",
		"1.0.0; curl evil.com",
		"1.0.0\nmalicious",
		"../../../tmp/evil",
		"",
		"abc",
		"1.0.0-beta",
		"v1.0.0",
	}
	for _, v := range invalid {
		if config.IsValidVersion(v) {
			t.Errorf("IsValidVersion(%q) = true, want false", v)
		}
	}
}

// #2: Language name validation rejects injection
func TestLangNameValidation(t *testing.T) {
	valid := []string{"go", "python", "node", "ruby", "rust", "zig", "node-js"}
	for _, n := range valid {
		if !config.IsValidLangName(n) {
			t.Errorf("IsValidLangName(%q) = false, want true", n)
		}
	}

	invalid := []string{
		"../etc",
		"lang; rm -rf /",
		"lang\nname",
		"",
		".hidden",
		"lang name",
		"$(whoami)",
	}
	for _, n := range invalid {
		if config.IsValidLangName(n) {
			t.Errorf("IsValidLangName(%q) = true, want false", n)
		}
	}
}

// #3: Bin path validation rejects traversal and injection
func TestBinPathValidation(t *testing.T) {
	valid := []string{"bin/go", "bin/python3", "zig", "bin/rustc", "go/bin/go"}
	for _, p := range valid {
		if !config.IsValidBinPath(p) {
			t.Errorf("IsValidBinPath(%q) = false, want true", p)
		}
	}

	invalid := []string{
		"../../../etc/passwd",
		"bin/../../evil",
		"bin/go; rm -rf /",
		"",
		"bin/go\nmalicious",
	}
	for _, p := range invalid {
		if config.IsValidBinPath(p) {
			t.Errorf("IsValidBinPath(%q) = true, want false", p)
		}
	}
}

// #4: ParseVernFile rejects path traversal versions
func TestParseVernFileRejectsTraversal(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), ".vern")

	// Path traversal in version
	os.WriteFile(tmp, []byte("python ../../tmp/evil"), 0644)
	_, _, err := config.ParseVernFile(tmp)
	if err == nil {
		t.Error("ParseVernFile should reject path traversal version")
	}

	// Shell injection in version
	os.WriteFile(tmp, []byte("python 1.0.0; curl evil.com"), 0644)
	_, _, err = config.ParseVernFile(tmp)
	if err == nil {
		t.Error("ParseVernFile should reject shell injection in version")
	}

	// Invalid language name
	os.WriteFile(tmp, []byte("../evil 1.0.0"), 0644)
	_, _, err = config.ParseVernFile(tmp)
	if err == nil {
		t.Error("ParseVernFile should reject invalid language name")
	}

	// Valid file still works
	os.WriteFile(tmp, []byte("python 3.13.13"), 0644)
	lang, ver, err := config.ParseVernFile(tmp)
	if err != nil {
		t.Fatalf("valid .vern file rejected: %v", err)
	}
	if lang != "python" || ver != "3.13.13" {
		t.Errorf("got %s %s, want python 3.13.13", lang, ver)
	}
}

// #5: Tar extraction rejects symlinks
func TestTarSymlinkRejection(t *testing.T) {
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	tw.WriteHeader(&tar.Header{
		Name:     "evil-link",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	})
	tw.Close()

	destDir := t.TempDir()
	tr := tar.NewReader(&tarBuf)
	err := install.ExtractTar(tr, destDir)
	if err == nil {
		t.Error("ExtractTar should reject symlink entries")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("error should mention symlink, got: %v", err)
	}
}

// #6: Tar extraction strips setuid bits
func TestTarSetuidStripped(t *testing.T) {
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	content := []byte("#!/bin/sh\necho hello")
	tw.WriteHeader(&tar.Header{
		Name:     "setuid-binary",
		Size:     int64(len(content)),
		Mode:     04755, // setuid
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)
	tw.Close()

	destDir := t.TempDir()
	tr := tar.NewReader(&tarBuf)
	err := install.ExtractTar(tr, destDir)
	if err != nil {
		t.Fatalf("ExtractTar failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(destDir, "setuid-binary"))
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	mode := info.Mode()
	if mode&os.ModeSetuid != 0 {
		t.Errorf("setuid bit should be stripped, got mode %o", mode)
	}
	if mode.Perm() != 0755 {
		t.Errorf("expected 0755, got %o", mode.Perm())
	}
}

// #7: Tar extraction rejects path traversal
func TestTarPathTraversal(t *testing.T) {
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	content := []byte("malicious")
	tw.WriteHeader(&tar.Header{
		Name:     "../../../tmp/evil",
		Size:     int64(len(content)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)
	tw.Close()

	destDir := t.TempDir()
	tr := tar.NewReader(&tarBuf)
	err := install.ExtractTar(tr, destDir)
	if err == nil {
		t.Error("ExtractTar should reject path traversal")
	}
	if !strings.Contains(err.Error(), "traversal") {
		t.Errorf("error should mention traversal, got: %v", err)
	}
}
