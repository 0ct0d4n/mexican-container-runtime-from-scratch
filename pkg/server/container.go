package server

import (
	"axolotl/pkg/model"
	"axolotl/pkg/server/rootfs"
	"encoding/json"
	"github.com/DanyelMorales/style"
	"log"
	"os"
	"os/exec"
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
	style.SuccessfulActionF("Image installed successfully, about to create the child process: ", image)
	cmd := exec.Command("/proc/self/exe", "init-container")
	stdin, _ := cmd.StdinPipe()
	go func() {
		json.NewEncoder(stdin).Encode(image)
		stdin.Close()
	}()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
