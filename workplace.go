package slck

import (
	"context"
	"fmt"
)

func NewWorkplace(cmds <-chan Command) *workplace {
	return &workplace{
		channels: make(map[string]channel),
		members:  make(map[username]member),
		cmds:     cmds,
	}
}

type workplace struct {
	channels map[string]channel
	members  map[username]member
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
				w.register(cmd.target)
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
			case messageInChannelCmd:
				w.sendMessageInChannel(cmd.sender, cmd.channel, cmd.body)
			case directMessageCmd:
				w.sendDirectMessage(cmd.sender, cmd.receipient, cmd.body)
			}
		}
	}
}

func (w *workplace) register(m member) {
	if _, ok := w.members[m.name]; ok {
		w.err(m, fmt.Sprintf("%s username is already taken", m.name))
		return
	}

	w.members[m.name] = m
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

func (w *workplace) sendDirectMessage(s Client, r string, body []byte) {
	m, ok := w.members[r]
	if !ok {
		return
	}

	fmt.Fprintf(m.conn, "%s: %s\n", s.username, body)
}

func (w *workplace) err(subject member, body string) {
	msg := msg{
		sender:  memberWorkspalce,
		subject: subject,
	}
	fmt.Fprintf(msg, "%s %s", commandErr, body)
}
