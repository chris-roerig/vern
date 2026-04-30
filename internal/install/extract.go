package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtractArchive extracts a tar.gz or tar.xz archive to destDir.
func ExtractArchive(filePath, destDir, extractType string) error {
	switch extractType {
	case "tar.gz":
		return extractTarGz(filePath, destDir)
	case "tar.xz":
		return extractTarXz(filePath, destDir)
	default:
		return fmt.Errorf("unsupported archive type: %s", extractType)
	}
}

func extractTarGz(filePath, destDir string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	return ExtractTar(tr, destDir)
}

func extractTarXz(filePath, destDir string) error {
	cmd := exec.Command("xz", "-dc", filePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create xz pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xz: %w", err)
	}

	tr := tar.NewReader(stdout)
	tarErr := ExtractTar(tr, destDir)

	if waitErr := cmd.Wait(); waitErr != nil {
		if tarErr != nil {
			return fmt.Errorf("xz failed: %w (tar error: %v)", waitErr, tarErr)
		}
		return fmt.Errorf("xz failed: %w", waitErr)
	}
	return tarErr
}

// ExtractTar extracts a tar stream to destDir, stripping the top-level directory.
// Rejects symlinks and hardlinks for security. Strips setuid/setgid bits.
func ExtractTar(tr *tar.Reader, destDir string) error {
	prefix := ""
	first := true

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Detect common prefix (top-level dir)
		if first {
			parts := strings.SplitN(hdr.Name, "/", 2)
			if len(parts) == 2 {
				prefix = parts[0] + "/"
			}
			first = false
		}

		// Strip prefix
		name := hdr.Name
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		if name == "" {
			continue
		}

		target := filepath.Join(destDir, name)

		// Path traversal protection
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry attempts path traversal: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("tar contains symlink entry (rejected for security): %s", hdr.Name)
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			wf, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(wf, tr); err != nil {
				wf.Close()
				return err
			}
			wf.Close()
			// Strip setuid/setgid bits for security
			if err := os.Chmod(target, os.FileMode(hdr.Mode)&0755); err != nil {
				return err
			}
		}
	}

	return nil
}
