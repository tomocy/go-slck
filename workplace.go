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
			sender:  memberWorkspalce,
			subject: m,
		}
		fmt.Fprint(msg, n)
	}
}

func (w *workplace) listMembers(m member) {
	for n := range w.members {
		msg := msg{
			sender:  memberWorkspalce,
			subject: m,
		}
		fmt.Fprint(msg, n)
	}
}

func (w *workplace) sendMessageInChannel(sender member, chName channelName, body []byte) {
	ch, ok := w.channels[chName]
	if !ok {
		return
	}

	ch.broadcast(sender, body)
}

func (w *workplace) sendDirectMessage(sender member, recipientName username, body []byte) {
	recipient, ok := w.members[recipientName]
	if !ok {
		return
	}

	msg := msg{
		sender:  sender,
		subject: recipient,
	}
	msg.Write(body)
}

func (w *workplace) err(subject member, body string) {
	msg := msg{
		sender:  memberWorkspalce,
		subject: subject,
	}
	fmt.Fprintf(msg, "%s %s", commandErr, body)
}
