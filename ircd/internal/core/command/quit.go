// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("QUIT", Quit)
}

func Quit(s *core.Session, m *irc.Message) {
	reason := m.Trailing
	quitMsg := &irc.Message{
		Prefix:   s.Prefix(),
		Command:  "QUIT",
		Trailing: reason,
	}

	// Broadcast to all channels user is in
	s.Srv.ChansMu.RLock()
	for _, ch := range s.Srv.Channels {
		ch.Mutex.RLock()
		if _, ok := ch.Users[s]; ok {
			ch.Mutex.RUnlock()
			ch.Broadcast(s, quitMsg)
			ch.Mutex.Lock()
			delete(ch.Users, s)
			ch.Mutex.Unlock()
		} else {
			ch.Mutex.RUnlock()
		}
	}
	s.Srv.ChansMu.RUnlock()

	// Remove from user map
	s.Srv.UsersMu.Lock()
	delete(s.Srv.Users, strings.ToLower(s.Nick))
	s.Srv.UsersMu.Unlock()

	s.Shutdown(m.Trailing)
}
