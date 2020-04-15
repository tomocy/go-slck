package slck

import (
	"fmt"
	"io"
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

func (c client) handle(cmd rawCommand) {
	switch cmd.kind {
	default:
		c.err(fmt.Sprintf("unknown command: %s", cmd.kind))
	}
}

func (c client) err(msg string) {
	fmt.Fprintf(c.conn, "%s %s\n", commandErr, msg)
}

type command struct {
	kind       commandKind
	sender     string
	receipient string
	body       []byte
}

type rawCommand struct {
	kind commandKind
	args []byte
}

func (c *rawCommand) Scan(state fmt.ScanState, _ rune) error {
	if _, err := fmt.Fscanf(state, "%s", &c.kind); err != nil {
		return fmt.Errorf("failed to parse command: %w", err)
	}
	args, err := state.Token(true, func(r rune) bool {
		return true
	})
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to parse args: %w", err)
	}
	c.args = make([]byte, len(args))
	copy(c.args, args)

	return nil
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
