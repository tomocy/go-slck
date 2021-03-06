package slck

import (
	"fmt"
	"net"
)

type channel struct {
	name    channelName
	members map[username]member
}

func (c channel) broadcast(from member, body []byte) {
	body = []byte(fmt.Sprintf("%s: %s\n", from, body))

	for _, m := range c.members {
		msg := msg{
			from: from,
			to:   m,
		}
		msg.Write(body)
	}
}

func (c *channel) join(m member) {
	c.members[m.name] = m
}

func (c *channel) leave(m member) {
	delete(c.members, m.name)
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
	from, to member
}

func (m msg) Write(body []byte) (int, error) {
	return fmt.Fprintf(m.to, "%s: %s\n", m.from.name, body)
}

var (
	memberWorkspalce = member{
		name: "workplace",
	}
)

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
