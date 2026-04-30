package tests

import (
	"testing"

	"github.com/chris-roerig/vern/internal/install"
)

func TestArchAlt(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"amd64", "x64"},
		{"arm64", "arm64"},
		{"386", "386"},
	}
	for _, tt := range tests {
		if got := install.ArchAlt(tt.input); got != tt.want {
			t.Errorf("ArchAlt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestArchGNU(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"amd64", "x86_64"},
		{"arm64", "aarch64"},
		{"386", "386"},
	}
	for _, tt := range tests {
		if got := install.ArchGNU(tt.input); got != tt.want {
			t.Errorf("ArchGNU(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestOsAlt(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"darwin", "macos"},
		{"linux", "linux"},
		{"windows", "windows"},
	}
	for _, tt := range tests {
		if got := install.OsAlt(tt.input); got != tt.want {
			t.Errorf("OsAlt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRustTarget(t *testing.T) {
	tests := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "aarch64-apple-darwin"},
		{"darwin", "amd64", "x86_64-apple-darwin"},
		{"linux", "amd64", "x86_64-unknown-linux-gnu"},
		{"linux", "arm64", "aarch64-unknown-linux-gnu"},
	}
	for _, tt := range tests {
		if got := install.RustTarget(tt.goos, tt.goarch); got != tt.want {
			t.Errorf("RustTarget(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}
