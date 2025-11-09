package client

import (
	"axolotl/pkg/model"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Req     *RunRequest
	Session *ssh.Session
}

type RunRequest struct {
	Rootfs    string                 `json:"rootfs"`
	Command   []string               `json:"command"`
	Namespace *model.NamespaceConfig `json:"cgroup"`
}

func NewClient(rootfs string, command []string, namespace *model.NamespaceConfig, Session *ssh.Session) *Client {
	return &Client{
		Req:     NewRunRequest(rootfs, command, namespace),
		Session: Session,
	}
}

func NewRunRequest(rootfs string, command []string, namespace *model.NamespaceConfig) *RunRequest {
	return &RunRequest{Rootfs: rootfs, Command: command, Namespace: namespace}
}

func NewCgroup(memory, cpu, pids int64) model.CgroupNamespace {
	return model.CgroupNamespace{
		MemoryMax: memory,
		CPUMax:    cpu,
		PidsMax:   pids,
	}
}
