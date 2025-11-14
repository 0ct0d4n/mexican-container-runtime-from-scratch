package main

import (
	"axolotl/pkg/server/rootfs"
	"encoding/json"
	"fmt"
	"os"
)

func runContainerInit() error {
	var image *rootfs.RootFSInstallationConfig
	if err := json.NewDecoder(os.Stdin).Decode(&image); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return image.Mount()
}
