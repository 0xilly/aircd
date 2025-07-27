// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("TOPIC", Topic)
}

func Topic(s *core.Session, m *irc.Message) {
	// Need at least channel name
	if len(m.Params) < 1 {
		s.Numeric(461, "TOPIC", "Not enough parameters")
		return
	}
	chName := m.Params[0]

	// Lookup channel
	ch := s.Srv.GetChannel(chName)
	if ch == nil {
		s.Numeric(403, chName, "No such channel")
		return
	}

	// View topic
	if len(m.Params) == 1 {
		topic := ch.Topic()
		if topic == "" {
			s.Numeric(331, s.Nick, chName, "No topic is set") // RPL_NOTOPIC
		} else {
			s.Numeric(332, s.Nick, chName, topic) // RPL_TOPIC
		}
		return
	}

	// Set topic (must be member; enforcement later)
	newTopic := m.Trailing
	ch.SetTopic(newTopic)

	// Broadcast new topic
	msg := &irc.Message{
		Prefix:   s.Prefix(),
		Command:  "TOPIC",
		Params:   []string{chName},
		Trailing: newTopic,
	}
	// Echo to setter and others
	s.Send(msg)
	ch.Broadcast(s, msg)
}
