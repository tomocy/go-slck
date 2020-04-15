package slck

import (
	"context"
	"fmt"
	"io"
	"net"
)

type channel struct {
	name    string
	members map[string]Client
}

func (c channel) broadcast(sender string, body []byte) {
	msg := []byte(fmt.Sprintf("%s: %s", sender, body))
	for m := range c.members {
		m.conn.Write(msg)
	}
}

func NewClient(conn net.Conn, registered chan<- Client) *Client {
	return &Client{
		conn: conn,
	}
}

type Client struct {
	conn       net.Conn
	username   string
	registered chan<- Client
}

func (c Client) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.err(fmt.Sprint(ctx.Err()))
			return
		default:
			var cmd rawCommand
			if _, err := fmt.Fscanf(c.conn, "%v\n", &cmd); err != nil {
				c.err(fmt.Sprintf("failed to scan command: %s", err))
				continue
			}

			c.handle(cmd)
		}
	}
}

func (c Client) handle(cmd rawCommand) {
	switch cmd.kind {
	default:
		c.err(fmt.Sprintf("unknown command: %s", cmd.kind))
	}
}

func (c Client) err(msg string) {
	c.printf("%s %s\n", commandErr, msg)
}

func (c Client) printf(format string, as ...interface{}) {
	fmt.Fprintf(c.conn, format, as...)
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
		return r != '\n'
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
