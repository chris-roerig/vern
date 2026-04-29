package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chris/vern/internal/install"
)

func TestGetInstalledVersionsSorted(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create fake install directories in wrong order
	langDir := filepath.Join(tmpHome, ".local", "share", "vern", "installs", "python")
	for _, v := range []string{"3.9.1", "3.13.13", "3.11.2", "3.10.0"} {
		os.MkdirAll(filepath.Join(langDir, v), 0755)
	}

	versions, err := install.GetInstalledVersions("python")
	if err != nil {
		t.Fatalf("GetInstalledVersions() error: %v", err)
	}

	expected := []string{"3.9.1", "3.10.0", "3.11.2", "3.13.13"}
	if len(versions) != len(expected) {
		t.Fatalf("got %d versions, want %d", len(versions), len(expected))
	}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("versions[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestGetInstalledVersionsEmpty(t *testing.T) {
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", origHome)

	versions, err := install.GetInstalledVersions("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected empty, got %v", versions)
	}
}

func TestGetInstalledLanguages(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	base := filepath.Join(tmpHome, ".local", "share", "vern", "installs")
	os.MkdirAll(filepath.Join(base, "go", "1.26.2"), 0755)
	os.MkdirAll(filepath.Join(base, "python", "3.13.13"), 0755)

	langs, err := install.GetInstalledLanguages()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(langs) != 2 {
		t.Errorf("got %d languages, want 2", len(langs))
	}
	if len(langs["go"]) != 1 || langs["go"][0] != "1.26.2" {
		t.Errorf("go versions = %v", langs["go"])
	}
}
