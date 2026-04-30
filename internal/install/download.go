// Package install handles downloading, extracting, building, and managing
// programming language installations for vern.
package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// HTTPClient is the shared HTTP client used for all downloads and API calls.
var HTTPClient = &http.Client{Timeout: 5 * time.Minute}

// Options controls install behavior.
type Options struct {
	Verbose bool
}

// progressReader wraps an io.Reader to display download progress.
type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	pct := float64(pr.current) / float64(pr.total) * 100
	mb := float64(pr.current) / 1024 / 1024
	totalMB := float64(pr.total) / 1024 / 1024
	fmt.Fprintf(os.Stdout, "\r  %.1f/%.1f MB (%.0f%%)", mb, totalMB, pct)
	return n, err
}

// DownloadFile downloads a URL to a temporary file, showing a progress bar.
// Returns the path to the temp file. Caller is responsible for cleanup.
func DownloadFile(url string) (string, error) {
	resp, err := HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "vern-download-*")
	if err != nil {
		return "", err
	}

	reader := io.LimitReader(resp.Body, 500*1024*1024)
	if resp.ContentLength > 0 {
		reader = &progressReader{reader: reader, total: resp.ContentLength}
	}

	_, err = io.Copy(tmpFile, reader)
	if resp.ContentLength > 0 {
		fmt.Print("\n")
	}
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// runCommand executes a shell command in the given directory.
func runCommand(dir, cmdStr string, opts Options) error {
	cmd := newCommand(dir, cmdStr)
	if opts.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func newCommand(dir, cmdStr string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir
	return cmd
}
