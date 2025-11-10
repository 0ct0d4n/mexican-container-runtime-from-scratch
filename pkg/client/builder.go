package client

import (
	"axolotl/pkg/model"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Req     *model.RunRequest
	Session *ssh.Session
}

func NewClient(rootfs string, command []string, namespace *model.NamespaceConfig, Session *ssh.Session) *Client {
	return &Client{
		Req:     NewRunRequest(rootfs, command, namespace),
		Session: Session,
	}
}

func NewRunRequest(rootfs string, command []string, namespace *model.NamespaceConfig) *model.RunRequest {
	return &model.RunRequest{Rootfs: rootfs, Command: command, Namespace: namespace}
}
