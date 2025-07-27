// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("KICK", Kick)
}

// KICK <channel> <user> [<reason>]
func Kick(s *core.Session, m *irc.Message) {
	// Need at least channel and target user
	if len(m.Params) < 2 {
		s.Numeric(461, "KICK", "Not enough parameters")
		return
	}
	chName := m.Params[0]
	targetNick := m.Params[1]
	reason := m.Trailing

	// Lookup channel
	ch := s.Srv.GetChannel(chName)
	if ch == nil {
		s.Numeric(403, chName, "No such channel")
		return
	}

	// Lookup target session
	s.Srv.UsersMu.RLock()
	target := s.Srv.Users[strings.ToLower(targetNick)]
	s.Srv.UsersMu.RUnlock()
	if target == nil {
		s.Numeric(441, chName, targetNick, "They aren't on that channel") // ERR_USERNOTINCHANNEL
		return
	}

	// Check membership
	ch.Mutex.RLock()
	_, member := ch.Users[target]
	ch.Mutex.RUnlock()
	if !member {
		s.Numeric(441, chName, targetNick, "They aren't on that channel")
		return
	}

	// Remove the user
	ch.Remove(target)

	// Build and send the KICK message
	msg := &irc.Message{
		Prefix:   s.Prefix(),
		Command:  "KICK",
		Params:   []string{chName, targetNick},
		Trailing: reason,
	}
	// Echo to kicker
	s.Send(msg)
	// Broadcast to remaining members
	ch.Broadcast(s, msg)
}
