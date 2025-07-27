// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("PRIVMSG", PrivMsg)
}

func PrivMsg(s *core.Session, m *irc.Message) {
	if len(m.Params) == 0 || m.Trailing == "" {
		s.Numeric(412, "No text to send") // ERR_NOTEXTTOSEND
		return
	}
	target := m.Params[0]
	text := m.Trailing

	msg := &irc.Message{
		Prefix:   s.Prefix(),
		Command:  "PRIVMSG",
		Params:   []string{target},
		Trailing: text,
	}

	if isChannel(target) {
		low := strings.ToLower(target)
		s.Srv.ChansMu.RLock()
		ch, ok := s.Srv.Channels[low]
		s.Srv.ChansMu.RUnlock()
		if !ok {
			s.Numeric(403, target, "No such channel")
			return
		}
		ch.Broadcast(s, msg)
	} else {
		// User message
		low := strings.ToLower(target)
		s.Srv.UsersMu.RLock()
		dst, ok := s.Srv.Users[low]
		s.Srv.UsersMu.RUnlock()
		if !ok {
			s.Numeric(401, target, "No such nick") // ERR_NOSUCHNICK
			return
		}
		dst.Send(msg)
	}
}
