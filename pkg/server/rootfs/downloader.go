package rootfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func DownloadRootFS(url, diskPath string) error {
	if err := os.MkdirAll(getDir(diskPath), 0755); err != nil {
		return fmt.Errorf("error creando directorio destino: %w", err)
	}

	// Hacer request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error haciendo GET: %w", err)
	}
	defer resp.Body.Close()

	// Validar status code
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("servidor devolvió código HTTP %d", resp.StatusCode)
	}

	// Crear archivo destino
	out, err := os.Create(diskPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear %s: %w", diskPath, err)
	}
	defer out.Close()

	// Copia en streaming, eficiente para archivos grandes
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return nil
}

func getDir(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return "."
	}
	return path[:i]
}

func FilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
