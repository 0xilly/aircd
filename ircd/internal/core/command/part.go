// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("PART", Part)
}

func Part(s *core.Session, m *irc.Message) {
	if len(m.Params) == 0 {
		s.Numeric(461, "PART", "Not enough parameters")
		return
	}
	names := strings.Split(m.Params[0], ",")
	reason := m.Trailing

	for _, name := range names {
		low := strings.ToLower(name)
		s.Srv.ChansMu.RLock()
		ch, ok := s.Srv.Channels[low]
		s.Srv.ChansMu.RUnlock()
		if !ok {
			s.Numeric(403, name, "No such channel") // ERR_NOSUCHCHANNEL
			continue
		}

		partMsg := &irc.Message{
			Prefix:   s.Prefix(),
			Command:  "PART",
			Params:   []string{name},
			Trailing: reason,
		}

		// Remove & broadcast
		ch.Mutex.Lock()
		delete(ch.Users, s)
		ch.Mutex.Unlock()

		s.Send(partMsg)
		ch.Broadcast(s, partMsg)
	}
}
