package server

import (
	"axolotl/pkg/model"
	"axolotl/pkg/server/rootfs"
	"github.com/DanyelMorales/style"
	"log"
)

const DefaultImagesPath = "/var/axolotl/images/"

func StartContainer(err error, payload *model.RunRequest) error {
	cfgGroup, err := createCgroup(payload.Namespace)
	if err != nil {
		log.Printf("[CGROUP] Error: failed to create instance: %v", err)
		return err
	}
	err = cfgGroup.LimitResources()
	if err != nil {
		log.Printf("[CGROUP] Error: failed to limit the resources on the instance: %v", err)
		return err
	}

	image, err := rootfs.InstallImage(DefaultImagesPath, payload.Namespace)
	if err != nil {
		return err
	}
	style.SuccessfulActionF("Operation completed successfully %v", image)
	return image.Mount()
}
