package client

import (
	axolotl "axolotl/pkg/client"
	"axolotl/pkg/model"
	"github.com/docker/go-units"
	"golang.org/x/crypto/ssh"
	"log"
)

func main() {
	config := &ssh.ClientConfig{
		User:            "axolotl",
		Auth:            []ssh.AuthMethod{ssh.Password("axolotl")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "192.168.64.2:2222", config)
	if err != nil {
		log.Fatalf("Error connecting to daemon %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Error creating SSH session: %v", err)
	}
	defer session.Close()

	ns := model.NamespaceConfig{
		Cgroup: &model.CgroupNamespace{
			MemoryMax: 100 * units.MB, // 100 MB
			CPUMax:    0.5,            // 50% de CPU
			PidsMax:   100,
			Path:      "test",
		},
	}

	c := axolotl.NewClient("", nil, &ns, session)
	if err := c.Create(); err != nil {
		log.Fatalf("Error creating container: %v", err)
	}
}
