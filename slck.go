package slck

import (
	"context"
	"fmt"
	"io"
	"net"
)

func NewWorkplace(cmds <-chan Command) *workplace {
	return &workplace{
		members: make(map[string]Client),
		cmds:    cmds,
	}
}

type workplace struct {
	channels map[string]channel
	members  map[string]Client
	cmds     <-chan Command
}

func (w workplace) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-w.cmds:
			switch cmd := cmd.(type) {
			case registerCmd:
				w.register(cmd.client)
			}
		}
	}
}

func (w *workplace) register(c Client) {
	if _, ok := w.members[c.username]; ok {
		c.err(fmt.Sprintf("%s username is already taken", c.username))
		return
	}

	w.members[c.username] = c
}

func (w *workplace) delete(cli Client) {
	delete(w.members, cli.username)
	for _, c := range w.channels {
		delete(c.members, cli.username)
	}
}

type channel struct {
	name    string
	members map[string]Client
}

func (c channel) broadcast(sender string, body []byte) {
	msg := []byte(fmt.Sprintf("%s: %s", sender, body))
	for _, m := range c.members {
		m.conn.Write(msg)
	}
}

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
		}
	case commandDelete:
		if err := c.delete(); err != nil {
			c.err(fmt.Sprintf("failed to delete: %s", err))
		}
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
	c.username = ""

	return nil
}

func (c Client) err(msg string) {
	c.printf("%s %s\n", commandErr, msg)
}

func (c Client) printf(format string, as ...interface{}) {
	fmt.Fprintf(c.conn, format, as...)
}

type Command interface {
	command()
}

type rawCmd struct {
	kind cmdKind
	args []byte
}

func (c *rawCmd) Scan(state fmt.ScanState, _ rune) error {
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

type cmdKind string

const (
	commandRegister cmdKind = "REGISTER"
	commandDelete   cmdKind = "DELETE"
	commandJoin     cmdKind = "JOIN"
	commandLeave    cmdKind = "LEAVE"
	commandChannels cmdKind = "CHANNELS"
	commandUsers    cmdKind = "USERS"
	commandMessage  cmdKind = "MESSAGE"
	commandOK       cmdKind = "OK"
	commandErr      cmdKind = "ERR"
)

type registerCmd struct {
	client Client
}

func (c registerCmd) command() {}
