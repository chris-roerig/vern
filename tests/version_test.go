package tests

import (
	"testing"

	"github.com/chris/vern/internal/version"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"1.21.0", 1, 21, 0, false},
		{"3.13.13", 3, 13, 13, false},
		{"0.14.0", 0, 14, 0, false},
		{"1.95.0", 1, 95, 0, false},
		{"21.0.11", 21, 0, 11, false},
		{"1.0", 0, 0, 0, true},
		{"abc", 0, 0, 0, true},
		{"1.2.3.4", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := version.ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVersion(%q) unexpected error: %v", tt.input, err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
				t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
			}
			if v.Full != tt.input {
				t.Errorf("ParseVersion(%q).Full = %q", tt.input, v.Full)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int // -1, 0, 1 (sign)
	}{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.2.0", "1.3.0", -1},
		{"1.2.3", "1.2.4", -1},
		{"3.13.13", "3.13.12", 1},
		{"0.14.0", "0.14.1", -1},
		{"1.95.0", "1.87.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			va, _ := version.ParseVersion(tt.a)
			vb, _ := version.ParseVersion(tt.b)
			got := va.Compare(vb)
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("Compare(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
