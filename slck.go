package slck

import "net"

type channel struct {
	name string
}

type client struct {
	conn net.Conn
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
