package slck

import (
	"fmt"
	"net"
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

type msg struct {
	sender  member
	subject member
}

func (m msg) Write(body []byte) (int, error) {
	return fmt.Fprintf(m.subject, "%s: %s\n", m.sender.name, body)
}

type member struct {
	name username
	conn net.Conn
}

func (m member) Write(src []byte) (int, error) {
	return m.conn.Write(src)
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
