package slck

import (
	"context"
	"fmt"
	"io"
	"net"
)

func NewWorkplace(cmds <-chan Command) *workplace {
	return &workplace{
		channels: make(map[string]channel),
		members:  make(map[string]Client),
		cmds:     cmds,
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
			case deleteCmd:
				w.delete(cmd.client)
			case joinCmd:
				w.join(cmd.client, cmd.channel)
			case leaveCmd:
				w.leave(cmd.client, cmd.channel)
			case channelsCmd:
				w.listChannels(cmd.client)
			case membersCmd:
				w.listMembers(cmd.client)
			case messageInChannel:
				w.sendMessageInChannel(cmd.sender, cmd.channel, cmd.body)
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

func (w *workplace) join(c Client, chName string) {
	if _, ok := w.channels[chName]; !ok {
		w.channels[chName] = channel{
			name:    chName,
			members: make(map[string]Client),
		}
	}

	w.channels[chName].members[c.username] = c
}

func (w *workplace) leave(c Client, chName string) {
	if _, ok := w.channels[chName]; !ok {
		return
	}

	delete(w.channels[chName].members, c.username)
}

func (w *workplace) listChannels(c Client) {
	for n := range w.channels {
		fmt.Fprintln(c.conn, n)
	}
}

func (w *workplace) listMembers(c Client) {
	for n := range w.members {
		fmt.Fprintln(c.conn, n)
	}
}

func (w *workplace) sendMessageInChannel(s Client, chName string, body []byte) {
	ch, ok := w.channels[chName]
	if !ok {
		return
	}

	ch.broadcast(s.username, body)
}

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
	case commandJoin:
		if err := c.join(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to join: %s", err))
		}
	case commandLeave:
		if err := c.leave(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to leave: %s", err))
		}
	case commandChannels:
		if err := c.channels(); err != nil {
			c.err(fmt.Sprintf("failed to list channels: %s", err))
		}
	case commandMembers:
		if err := c.members(); err != nil {
			c.err(fmt.Sprintf("failed to list members: %s", err))
		}
	case commandMessage:
		if err := c.message(cmd.args); err != nil {
			c.err(fmt.Sprintf("failed to send message: %s", err))
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

	if cmd.target[0] != '#' {
		return fmt.Errorf("invalid target format: format should start either @ or #: %s", cmd.target)
	}

	ch := channelName(cmd.target)
	if err := ch.validate(); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	c.cmds <- messageInChannel{
		sender:  c,
		channel: string(ch),
		body:    cmd.body,
	}

	return nil
}

func (c Client) err(msg string) {
	c.printf("%s %s\n", commandErr, msg)
}

func (c Client) printf(format string, as ...interface{}) {
	fmt.Fprintf(c.conn, format, as...)
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
	commandMembers  cmdKind = "MEMBERS"
	commandMessage  cmdKind = "MESSAGE"
	commandOK       cmdKind = "OK"
	commandErr      cmdKind = "ERR"
)

type rawMessageCmd struct {
	target string
	len    int
	body   []byte
}

func (c *rawMessageCmd) Scan(state fmt.ScanState, _ rune) error {
	if _, err := fmt.Fscanf(state, "%s %d ", &c.target, &c.len); err != nil {
		return fmt.Errorf("failed to scan target or length: %w", err)
	}

	var n int
	body, err := state.Token(false, func(r rune) bool {
		defer func() { n++ }()
		return n < c.len
	})
	if err != nil {
		return fmt.Errorf("failed to parse body: %w", err)
	}
	c.body = make([]byte, len(body))
	copy(c.body, body)

	return nil
}

type registerCmd struct {
	client Client
}

func (c registerCmd) command() {}

type deleteCmd struct {
	client Client
}

func (c deleteCmd) command() {}

type joinCmd struct {
	client  Client
	channel string
}

func (c joinCmd) command() {}

type leaveCmd struct {
	client  Client
	channel string
}

func (c leaveCmd) command() {}

type channelsCmd struct {
	client Client
}

func (c channelsCmd) command() {}

type membersCmd struct {
	client Client
}

func (c membersCmd) command() {}

type messageInChannel struct {
	sender  Client
	channel string
	body    []byte
}

func (c messageInChannel) command() {}
