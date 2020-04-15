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
	conn net.Conn
	as   username
	cmds chan<- Command
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
	case cmdRegister:
		if err := c.register(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to register: %s", err))
			return
		}

		c.ok()
	case cmdDelete:
		if err := c.delete(); err != nil {
			c.err(fmt.Sprintf("failed to delete: %s", err))
			return
		}

		c.ok()
	case cmdJoin:
		if err := c.join(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to join: %s", err))
			return
		}

		c.ok()
	case cmdLeave:
		if err := c.leave(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to leave: %s", err))
			return
		}

		c.ok()
	case cmdChannels:
		if err := c.channels(); err != nil {
			c.err(fmt.Sprintf("failed to list channels: %s", err))
			return
		}

		c.ok()
	case cmdMembers:
		if err := c.members(); err != nil {
			c.err(fmt.Sprintf("failed to list members: %s", err))
			return
		}

		c.ok()
	case cmdSend:
		if err := c.message(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to send message: %s", err))
			return
		}

		c.ok()
	default:
		c.err(fmt.Sprintf("unknown command: %s", cmd.kind))
	}
}

func (c *Client) register(args []byte) error {
	name := username(args)
	if err := name.validate(); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	c.as = name

	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	c.cmds <- registerCmd{
		target: m,
	}

	return nil
}

func (c *Client) delete() error {
	m, _ := c.member()

	c.cmds <- deleteCmd{
		target: m,
	}

	c.as = ""

	return nil
}

func (c Client) join(args []byte) error {
	ch := channelName(args)
	if err := ch.validate(); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	c.cmds <- joinCmd{
		member:  m,
		channel: ch,
	}

	return nil
}

func (c Client) leave(args []byte) error {
	ch := channelName(args)
	if err := ch.validate(); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	c.cmds <- leaveCmd{
		member:  m,
		channel: ch,
	}

	return nil
}

func (c Client) channels() error {
	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	c.cmds <- channelsCmd{
		member: m,
	}

	return nil
}

func (c Client) members() error {
	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	c.cmds <- membersCmd{
		member: m,
	}

	return nil
}

func (c Client) message(args []byte) error {
	var cmd rawMessageCmd
	if _, err := fmt.Sscan(string(args), &cmd); err != nil {
		return fmt.Errorf("failed to scan command: %w", err)
	}

	m, err := c.member()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	if cmd.target[0] == '#' {
		ch := channelName(cmd.target)
		if err := ch.validate(); err != nil {
			return fmt.Errorf("invalid channel name: %w", err)
		}

		c.cmds <- messageInChannelCmd{
			from: m,
			in:   ch,
			body: cmd.body,
		}

		return nil
	}

	if cmd.target[0] == '@' {
		uname := username(cmd.target)
		if err := uname.validate(); err != nil {
			return fmt.Errorf("invalid channel name: %w", err)
		}

		c.cmds <- directMessageCmd{
			from: m,
			to:   uname,
			body: cmd.body,
		}

		return nil
	}

	return fmt.Errorf("invalid target format: format should start either @ or #: %s", cmd.target)
}

func (c Client) member() (member, error) {
	if c.as == "" {
		return member{}, fmt.Errorf("no session is available")
	}

	return member{
		name: c.as,
		conn: c.conn,
	}, nil
}

func (c Client) ok() {
	c.printf("%s\n", cmdOK)
}

func (c Client) err(msg string) {
	c.printf("%s %s\n", cmdErr, msg)
}

func (c Client) printf(format string, as ...interface{}) {
	fmt.Fprintf(c.conn, format, as...)
}
