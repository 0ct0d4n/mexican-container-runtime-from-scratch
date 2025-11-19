package util

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UntarRootFS(tarFile, destDir string) error {
	f, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("error abriendo tar: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("error leyendo gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error leyendo header tar: %w", err)
		}

		target := filepath.Join(destDir, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, os.FileMode(hdr.Mode))

		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)

			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("error creando archivo: %w", err)
			}

			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()

		case tar.TypeSymlink:
			os.Symlink(hdr.Linkname, target)
		}
	}

	return nil
}

func GenerateID() string {
	b := make([]byte, 8) // 64 bits = más que suficiente
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
