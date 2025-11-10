package client

import (
	"encoding/json"
)

func (r *Client) Create() error {
	data, _ := json.Marshal(r.Req)
	return r.SendPayload(data, "exec")
}

func (r *Client) SendPayload(data []byte, command string) error {
	stdin, _ := r.Session.StdinPipe()
	_, err := stdin.Write(data)
	if err != nil {
		return err
	}

	err = stdin.Close()
	if err != nil {
		return err
	}

	//err = r.Session.Run(command)
	//if err != nil {
	//	return err
	//}

	return nil
}
