package slck

import (
	"fmt"
	"io"
)

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
	target member
}

func (c registerCmd) command() {}

type deleteCmd struct {
	target member
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

type messageInChannelCmd struct {
	sender  Client
	channel string
	body    []byte
}

func (c messageInChannelCmd) command() {}

type directMessageCmd struct {
	sender     Client
	receipient string
	body       []byte
}

func (c directMessageCmd) command() {}
