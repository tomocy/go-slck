package slck

import (
	"fmt"
)

type channel struct {
	name    string
	members map[string]Client
}

func (c channel) broadcast(sender string, body []byte) {
	msg := []byte(fmt.Sprintf("%s: %s\n", sender, body))
	for _, m := range c.members {
		m.conn.Write(msg)
	}
}

type channelName string

func (n channelName) validate() error {
	if n == "" {
		return fmt.Errorf("name is empty")
	}
	if n[0] != '#' {
		return fmt.Errorf("name does not start with #")
	}
	if n[1:] == "" {
		return fmt.Errorf("name exluding # is empty")
	}

	return nil
}

type username string

func (n username) validate() error {
	if n == "" {
		return fmt.Errorf("name is empty")
	}
	if n[0] != '@' {
		return fmt.Errorf("name does not start with @")
	}
	if n[1:] == "" {
		return fmt.Errorf("name exluding @ is empty")
	}

	return nil
}
