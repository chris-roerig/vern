package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chris-roerig/vern/internal/config"
)

func TestParseVernFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantLang string
		wantVer  string
		wantErr  bool
	}{
		{"valid", "python 3.13.13", "python", "3.13.13", false},
		{"with newline", "go 1.26.2\n", "go", "1.26.2", false},
		{"with spaces", "  ruby 4.0.3  \n", "ruby", "4.0.3", false},
		{"empty", "", "", "", true},
		{"no version", "python", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), ".vern")
			os.WriteFile(tmp, []byte(tt.content), 0644)

			lang, ver, err := config.ParseVernFile(tmp)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if lang != tt.wantLang {
				t.Errorf("lang = %q, want %q", lang, tt.wantLang)
			}
			if ver != tt.wantVer {
				t.Errorf("ver = %q, want %q", ver, tt.wantVer)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	// With no defaults file, should return empty map
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", origHome)

	defaults, err := config.LoadDefaults()
	if err != nil {
		t.Fatalf("LoadDefaults() error: %v", err)
	}
	if len(defaults) != 0 {
		t.Errorf("expected empty defaults, got %v", defaults)
	}
}

func TestSaveAndLoadDefaults(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create the data directory
	dataDir := filepath.Join(tmpHome, ".local", "share", "vern")
	os.MkdirAll(dataDir, 0755)

	defaults := map[string]string{
		"go":     "1.26.2",
		"python": "3.13.13",
	}

	if err := config.SaveDefaults(defaults); err != nil {
		t.Fatalf("SaveDefaults() error: %v", err)
	}

	loaded, err := config.LoadDefaults()
	if err != nil {
		t.Fatalf("LoadDefaults() error: %v", err)
	}

	if loaded["go"] != "1.26.2" {
		t.Errorf("go = %q, want %q", loaded["go"], "1.26.2")
	}
	if loaded["python"] != "3.13.13" {
		t.Errorf("python = %q, want %q", loaded["python"], "3.13.13")
	}
}
