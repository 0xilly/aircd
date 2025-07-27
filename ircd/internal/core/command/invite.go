// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("INVITE", Invite)
}

func Invite(s *core.Session, m *irc.Message) {
	// Syntax: INVITE <nick> <channel>
	if len(m.Params) < 2 {
		s.Numeric(461, "INVITE", "Not enough parameters")
		return
	}
	targetNick := m.Params[0]
	chName := m.Params[1]

	// Lookup channel
	ch := s.Srv.GetChannel(chName)
	if ch == nil {
		s.Numeric(403, chName, "No such channel")
		return
	}

	// Must be on channel to invite
	ch.Mutex.RLock()
	_, member := ch.Users[s]
	ch.Mutex.RUnlock()
	if !member {
		s.Numeric(442, chName, "You're not on that channel") // ERR_NOTONCHANNEL
		return
	}

	// Lookup target session
	s.Srv.UsersMu.RLock()
	target := s.Srv.Users[strings.ToLower(targetNick)]
	s.Srv.UsersMu.RUnlock()
	if target == nil {
		s.Numeric(401, targetNick, "No such nick") // ERR_NOSUCHNICK
		return
	}

	// Send the INVITE to the target
	inviteMsg := &irc.Message{
		Prefix:  s.Prefix(),
		Command: "INVITE",
		Params:  []string{targetNick, chName},
	}
	target.Send(inviteMsg)

	// Acknowledge to the inviter: RPL_INVITING 341
	s.Numeric(341, s.Nick, targetNick, chName)
}
