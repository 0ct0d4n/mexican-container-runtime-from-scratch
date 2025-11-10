package client

import (
	"axolotl/pkg/model"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Req     *model.RunRequest
	Session *ssh.Session
}

func NewClient(command []string, namespace *model.NamespaceConfig, Session *ssh.Session) *Client {
	return &Client{
		Req:     NewRunRequest(command, namespace),
		Session: Session,
	}
}

func NewRunRequest(command []string, namespace *model.NamespaceConfig) *model.RunRequest {
	return &model.RunRequest{Command: command, Namespace: namespace}
}
