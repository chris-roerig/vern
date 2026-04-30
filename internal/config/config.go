package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	validVersion  = regexp.MustCompile(`^\d+(\.\d+){0,2}$`)
	validLangName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	validBinPath  = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
)

func IsValidVersion(v string) bool {
	return validVersion.MatchString(v)
}

func IsValidLangName(n string) bool {
	return validLangName.MatchString(n)
}

func IsValidBinPath(p string) bool {
	return validBinPath.MatchString(p) && !strings.Contains(p, "..")
}

type Language struct {
	Name          string        `yaml:"name"`
	BinaryName    string        `yaml:"binary_name"`
	VersionSource VersionSource `yaml:"version_source"`
	Install       Install       `yaml:"install"`
	Requires      *Requirement  `yaml:"requires,omitempty"`
}

type VersionSource struct {
	URL          string `yaml:"url"`
	VersionRegex string `yaml:"version_regex"`
}

type Install struct {
	DownloadTemplate string `yaml:"download_template"`
	ExtractType      string `yaml:"extract_type"`
	BinRelPath       string `yaml:"bin_rel_path"`
	BuildConfig      string `yaml:"build_config"`
	BuildCommand     string `yaml:"build_command"`
}

// Requirement defines a build dependency that must be satisfied before installing.
type Requirement struct {
	Binary           string `yaml:"binary"`
	MinVersion       string `yaml:"min_version"`
	BootstrapVersion string `yaml:"bootstrap_version"`
	WhenMajorGte     int    `yaml:"when_major_gte"`
}

type Config struct {
	Languages []Language `yaml:"languages"`
}

func LoadConfig() (*Config, error) {
	configPath := filepath.Join(ConfigDir(), "languages.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

func createDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	defaultCfg := Config{
		Languages: []Language{
			{
				Name:       "go",
				BinaryName: "go",
				VersionSource: VersionSource{
					URL:          "https://go.dev/dl/",
					VersionRegex: `go(\d+\.\d+\.\d+)\.linux-amd64\.tar\.gz`,
				},
				Install: Install{
					DownloadTemplate: "https://go.dev/dl/go{{.Version}}.{{.OS}}-{{.Arch}}.tar.gz",
					ExtractType:      "tar.gz",
					BinRelPath:       "bin/go",
				},
			},
			{
				Name:       "python",
				BinaryName: "python3",
				VersionSource: VersionSource{
					URL:          "https://www.python.org/ftp/python/",
					VersionRegex: `(\d+\.\d+\.\d+)/`,
				},
				Install: Install{
					DownloadTemplate: "https://www.python.org/ftp/python/{{.Version}}/Python-{{.Version}}.tgz",
					ExtractType:      "tar.gz",
					BinRelPath:       "bin/python3",
					BuildConfig:      `./configure --prefix={{.InstallDir}} $(uname -s | grep -q Darwin && echo "--build=$(uname -m | sed 's/arm64/aarch64/')-apple-darwin")`,
					BuildCommand:     `make -j$(sysctl -n hw.ncpu 2>/dev/null || nproc) && make install`,
				},
			},
			{
				Name:       "node",
				BinaryName: "node",
				VersionSource: VersionSource{
					URL:          "https://nodejs.org/dist/",
					VersionRegex: `v(\d+\.\d+\.\d+)/`,
				},
				Install: Install{
					DownloadTemplate: "https://nodejs.org/dist/v{{.Version}}/node-v{{.Version}}-{{.OS}}-{{.ArchAlt}}.tar.xz",
					ExtractType:      "tar.xz",
					BinRelPath:       "bin/node",
				},
			},
			{
				Name:       "ruby",
				BinaryName: "ruby",
				VersionSource: VersionSource{
					URL:          "https://www.ruby-lang.org/en/downloads/releases/",
					VersionRegex: `Ruby (\d+\.\d+\.\d+)`,
				},
				Install: Install{
					DownloadTemplate: "https://cache.ruby-lang.org/pub/ruby/{{.MajorMinor}}/ruby-{{.Version}}.tar.gz",
					ExtractType:      "tar.gz",
					BinRelPath:       "bin/ruby",
					BuildConfig:      `./configure --prefix={{.InstallDir}} $(uname -s | grep -q Darwin && echo "--build=$(uname -m | sed 's/arm64/aarch64/')-apple-darwin")`,
					BuildCommand:     `make -j$(sysctl -n hw.ncpu 2>/dev/null || nproc) && make install`,
				},
				Requires: &Requirement{
					Binary:           "ruby",
					MinVersion:       "3.1.0",
					BootstrapVersion: "3.4",
					WhenMajorGte:     4,
				},
			},
			{
				Name:       "rust",
				BinaryName: "rustc",
				VersionSource: VersionSource{
					URL:          "https://github.com/rust-lang/rust/tags",
					VersionRegex: `/releases/tag/(\d+\.\d+\.\d+)`,
				},
				Install: Install{
					DownloadTemplate: "https://static.rust-lang.org/dist/rust-{{.Version}}-{{.RustTarget}}.tar.xz",
					ExtractType:      "tar.xz",
					BinRelPath:       "bin/rustc",
					BuildConfig:      "./install.sh --prefix={{.InstallDir}}",
				},
			},
			{
				Name:       "zig",
				BinaryName: "zig",
				VersionSource: VersionSource{
					URL:          "https://ziglang.org/download/",
					VersionRegex: `zig-linux-x86_64-(\d+\.\d+\.\d+)\.tar\.xz`,
				},
				Install: Install{
					DownloadTemplate: "https://ziglang.org/download/{{.Version}}/zig-{{.OSAlt}}-{{.ArchGNU}}-{{.Version}}.tar.xz",
					ExtractType:      "tar.xz",
					BinRelPath:       "zig",
				},
			},
		},
	}
	data, err := yaml.Marshal(&defaultCfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LangsVersionPath() string {
	return filepath.Join(ConfigDir(), "langs_version")
}

func LoadLangsVersion() string {
	path := LangsVersionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(data))
}

func SaveLangsVersion(version string) error {
	return os.WriteFile(LangsVersionPath(), []byte(version), 0644)
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// ConfigDir returns the vern config directory.
// Respects VERN_HOME env var if set, otherwise uses ~/.config/vern.
func ConfigDir() string {
	if home := os.Getenv("VERN_HOME"); home != "" {
		return filepath.Join(home, "config")
	}
	return filepath.Join(homeDir(), ".config", "vern")
}

// DataDir returns the vern data directory.
// Respects VERN_HOME env var if set, otherwise uses ~/.local/share/vern.
func DataDir() string {
	if home := os.Getenv("VERN_HOME"); home != "" {
		return filepath.Join(home, "data")
	}
	return filepath.Join(homeDir(), ".local", "share", "vern")
}

func InstallsDir() string {
	return filepath.Join(DataDir(), "installs")
}

func LanguageInstallDir(lang, version string) string {
	return filepath.Join(InstallsDir(), lang, version)
}

func ShimsDir() string {
	return filepath.Join(DataDir(), "shims")
}

func DefaultsPath() string {
	return filepath.Join(DataDir(), "defaults.yaml")
}

func LoadDefaults() (map[string]string, error) {
	path := DefaultsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var defaults map[string]string
	if err := yaml.Unmarshal(data, &defaults); err != nil {
		return nil, err
	}
	if defaults == nil {
		defaults = make(map[string]string)
	}
	return defaults, nil
}

func SaveDefaults(defaults map[string]string) error {
	data, err := yaml.Marshal(defaults)
	if err != nil {
		return err
	}
	return os.WriteFile(DefaultsPath(), data, 0644)
}

func FindLocalVernFile() (string, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := startDir
	for {
		vernPath := filepath.Join(dir, ".vern")
		if _, err := os.Stat(vernPath); err == nil {
			return vernPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".vern file not found")
}

func ParseVernFile(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		return "", "", fmt.Errorf("empty .vern file")
	}
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid .vern format, expected 'language version'")
	}
	lang, ver := parts[0], strings.TrimSpace(parts[1])
	if !IsValidLangName(lang) {
		return "", "", fmt.Errorf("invalid language name in .vern: %q", lang)
	}
	if !IsValidVersion(ver) {
		return "", "", fmt.Errorf("invalid version in .vern: %q", ver)
	}
	return lang, ver, nil
}
