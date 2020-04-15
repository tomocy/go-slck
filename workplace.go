package slck

import (
	"context"
	"fmt"
)

func NewWorkplace(cmds <-chan Command) *workplace {
	return &workplace{
		channels: make(map[channelName]channel),
		members:  make(map[username]member),
		cmds:     cmds,
	}
}

type workplace struct {
	channels map[channelName]channel
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
				w.delete(cmd.target)
			case joinCmd:
				w.join(cmd.member, cmd.channel)
			case leaveCmd:
				w.leave(cmd.member, cmd.channel)
			case channelsCmd:
				w.listChannels(cmd.member)
			case membersCmd:
				w.listMembers(cmd.member)
			case messageInChannelCmd:
				w.sendMessageInChannel(cmd.from, cmd.in, cmd.body)
			case directMessageCmd:
				w.sendDirectMessage(cmd.from, cmd.to, cmd.body)
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

func (w *workplace) delete(m member) {
	delete(w.members, m.name)
	for _, c := range w.channels {
		c.leave(m)
	}
}

func (w *workplace) join(m member, chName channelName) {
	ch, ok := w.channels[chName]
	if !ok {
		ch = channel{
			name:    chName,
			members: make(map[username]member),
		}
		w.channels[chName] = ch
	}

	ch.join(m)
}

func (w *workplace) leave(m member, chName channelName) {
	ch, ok := w.channels[chName]
	if !ok {
		return
	}

	ch.leave(m)
}

func (w *workplace) listChannels(m member) {
	for n := range w.channels {
		msg := msg{
			from: memberWorkspalce,
			to:   m,
		}
		fmt.Fprint(msg, n)
	}
}

func (w *workplace) listMembers(m member) {
	for n := range w.members {
		msg := msg{
			from: memberWorkspalce,
			to:   m,
		}
		fmt.Fprint(msg, n)
	}
}

func (w *workplace) sendMessageInChannel(from member, chName channelName, body []byte) {
	ch, ok := w.channels[chName]
	if !ok {
		return
	}

	ch.broadcast(from, body)
}

func (w *workplace) sendDirectMessage(from member, recipientName username, body []byte) {
	to, ok := w.members[recipientName]
	if !ok {
		return
	}

	msg := msg{
		from: from,
		to:   to,
	}
	msg.Write(body)
}

func (w *workplace) err(to member, body string) {
	msg := msg{
		from: memberWorkspalce,
		to:   to,
	}
	fmt.Fprintf(msg, "%s %s", commandErr, body)
}
