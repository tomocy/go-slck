package slck

import (
	"fmt"
	"net"
)

type channel struct {
	name    string
	members map[string]client
}

func (c channel) broadcast(sender string, body []byte) {
	msg := []byte(fmt.Sprintf("%s: %s", sender, body))
	for m := range c.members {
		m.conn.Write(msg)
	}
}

type client struct {
	conn     net.Conn
	username string
}

type command struct {
	kind       commandKind
	sender     string
	receipient string
	body       []byte
}

type commandKind string

const (
	commandRegister commandKind = "REGISTER"
	commandJoin     commandKind = "JOIN"
	commandLeave    commandKind = "LEAVE"
	commandChannels commandKind = "CHANNELS"
	commandUsers    commandKind = "USERS"
	commandMessage  commandKind = "MESSAGE"
	commandOK       commandKind = "OK"
	commandErr      commandKind = "ERR"
)
