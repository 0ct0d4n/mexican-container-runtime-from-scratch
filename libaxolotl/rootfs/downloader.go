package rootfs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Download downloads a rootfs tarball from the given URL to the destination path
func Download(url, destPath string) error {
	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		log.Printf("[ROOTFS] File already exists, skipping download: %s", destPath)
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	log.Printf("[ROOTFS] Downloading %s -> %s", url, destPath)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Validate status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer out.Close()

	// Stream copy (efficient for large files)
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("[ROOTFS] Downloaded %d bytes to %s", bytesWritten, destPath)
	return nil
}

// FilenameFromURL extracts the filename from a URL
func FilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
