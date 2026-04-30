package install

// ArchAlt maps Go's GOARCH to Node.js-style architecture names.
func ArchAlt(goarch string) string {
	if goarch == "amd64" {
		return "x64"
	}
	return goarch
}

// ArchGNU maps Go's GOARCH to GNU-style architecture names.
func ArchGNU(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	}
	return goarch
}

// OsAlt maps Go's GOOS to alternative OS names (e.g., Zig uses "macos").
func OsAlt(goos string) string {
	if goos == "darwin" {
		return "macos"
	}
	return goos
}

// RustTarget returns the Rust target triple for the given OS and architecture.
func RustTarget(goos, goarch string) string {
	arch := ArchGNU(goarch)
	switch goos {
	case "darwin":
		return arch + "-apple-darwin"
	default:
		return arch + "-unknown-" + goos + "-gnu"
	}
}
