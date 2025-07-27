// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("JOIN", Join)
}

func Join(s *core.Session, m *irc.Message) {
	if len(m.Params) == 0 && m.Trailing == "" {
		s.Numeric(461, "JOIN", "Not enough parameters")
		return
	}
	targets := strings.Split(firstParam(m), ",")
	for _, name := range targets {
		if name == "" {
			continue
		}
		ch := s.Srv.GetOrCreateChannel(name)

		// Add member
		ch.Mutex.Lock()
		if ch.Users == nil {
			ch.Users = make(map[*core.Session]struct{})
		}
		ch.Users[s] = struct{}{}
		ch.Mutex.Unlock()

		joinMsg := &irc.Message{
			Prefix:  s.Prefix(),
			Command: "JOIN",
			Params:  []string{name},
		}
		// Echo to self & broadcast
		s.Send(joinMsg)
		ch.Broadcast(s, joinMsg)
	}
}
