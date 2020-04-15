package slck

import (
	"context"
	"fmt"
	"net"
)

func NewClient(conn net.Conn, cmds chan<- Command) *Client {
	return &Client{
		conn: conn,
		cmds: cmds,
	}
}

type Client struct {
	conn     net.Conn
	username string
	cmds     chan<- Command
}

func (c Client) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.err(fmt.Sprint(ctx.Err()))
			return
		default:
			var cmd rawCmd
			if _, err := fmt.Fscanf(c.conn, "%v\n", &cmd); err != nil {
				c.err(fmt.Sprintf("failed to scan command: %s", err))
				continue
			}

			c.handle(cmd)
		}
	}
}

func (c *Client) handle(cmd rawCmd) {
	switch cmd.kind {
	case commandRegister:
		if err := c.register(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to register: %s", err))
			return
		}

		c.ok()
	case commandDelete:
		if err := c.delete(); err != nil {
			c.err(fmt.Sprintf("failed to delete: %s", err))
		}

		c.ok()
	case commandJoin:
		if err := c.join(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to join: %s", err))
		}

		c.ok()
	case commandLeave:
		if err := c.leave(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to leave: %s", err))
		}

		c.ok()
	case commandChannels:
		if err := c.channels(); err != nil {
			c.err(fmt.Sprintf("failed to list channels: %s", err))
		}

		c.ok()
	case commandMembers:
		if err := c.members(); err != nil {
			c.err(fmt.Sprintf("failed to list members: %s", err))
		}

		c.ok()
	case commandMessage:
		if err := c.message(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to send message: %s", err))
		}

		c.ok()
	default:
		c.err(fmt.Sprintf("unknown command: %s", cmd.kind))
	}
}

func (c *Client) register(args []byte) error {
	name := string(args)

	if err := c.setUsername(name); err != nil {
		return err
	}

	c.cmds <- registerCmd{
		client: *c,
	}

	return nil
}

func (c *Client) setUsername(name string) error {
	if name == "" {
		return fmt.Errorf("username is empty")
	}
	if name[0] != '@' {
		return fmt.Errorf("username does not start with @")
	}
	if name[1:] == "" {
		return fmt.Errorf("username expluding @ is empty")
	}

	c.username = name

	return nil
}

func (c *Client) delete() error {
	c.cmds <- deleteCmd{
		client: *c,
	}

	c.username = ""

	return nil
}

func (c Client) join(args []byte) error {
	ch := channelName(args)
	if err := ch.validate(); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	c.cmds <- joinCmd{
		client:  c,
		channel: string(ch),
	}

	return nil
}

func (c Client) leave(args []byte) error {
	ch := channelName(args)
	if err := ch.validate(); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	c.cmds <- leaveCmd{
		client:  c,
		channel: string(ch),
	}

	return nil
}

func (c Client) channels() error {
	c.cmds <- channelsCmd{
		client: c,
	}

	return nil
}

func (c Client) members() error {
	c.cmds <- membersCmd{
		client: c,
	}

	return nil
}

func (c Client) message(args []byte) error {
	var cmd rawMessageCmd
	if _, err := fmt.Sscan(string(args), &cmd); err != nil {
		return fmt.Errorf("failed to scan command: %w", err)
	}

	if cmd.target[0] == '#' {
		ch := channelName(cmd.target)
		if err := ch.validate(); err != nil {
			return fmt.Errorf("invalid channel name: %w", err)
		}

		c.cmds <- messageInChannelCmd{
			sender:  c,
			channel: string(ch),
			body:    cmd.body,
		}

		return nil
	}

	if cmd.target[0] == '@' {
		uname := username(cmd.target)
		if err := uname.validate(); err != nil {
			return fmt.Errorf("invalid channel name: %w", err)
		}

		c.cmds <- directMessageCmd{
			sender:     c,
			receipient: string(uname),
			body:       cmd.body,
		}

		return nil
	}

	return fmt.Errorf("invalid target format: format should start either @ or #: %s", cmd.target)
}

func (c Client) ok() {
	c.printf("%s\n", commandOK)
}

func (c Client) err(msg string) {
	c.printf("%s %s\n", commandErr, msg)
}

func (c Client) printf(format string, as ...interface{}) {
	fmt.Fprintf(c.conn, format, as...)
}
