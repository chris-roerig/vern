package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Language struct {
	Name          string        `yaml:"name"`
	BinaryName    string        `yaml:"binary_name"`
	VersionSource VersionSource `yaml:"version_source"`
	Install       Install       `yaml:"install"`
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
					BuildConfig:      "./configure --prefix={{.InstallDir}}",
					BuildCommand:     "make -j$(nproc) && make install",
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
					URL:          "https://cache.ruby-lang.org/pub/ruby/",
					VersionRegex: `(\d+\.\d+\.\d+)/`,
				},
				Install: Install{
					DownloadTemplate: "https://cache.ruby-lang.org/pub/ruby/{{.MajorMinor}}/ruby-{{.Version}}.tar.gz",
					ExtractType:      "tar.gz",
					BinRelPath:       "bin/ruby",
					BuildConfig:      "./configure --prefix={{.InstallDir}}",
					BuildCommand:     "make -j$(nproc) && make install",
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

func ConfigDir() string {
	return filepath.Join(homeDir(), ".config", "vern")
}

func DataDir() string {
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
	return parts[0], parts[1], nil
}
