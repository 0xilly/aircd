// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("WHO", Who)
}

// WHO: WHO <channel|nick>
func Who(s *core.Session, m *irc.Message) {
	// Need a target
	if len(m.Params) < 1 {
		s.Numeric(461, "WHO", "Not enough parameters")
		return
	}
	target := m.Params[0]

	var sessions []*core.Session
	if strings.HasPrefix(target, "#") {
		// WHO #channel
		ch := s.Srv.GetChannel(target)
		if ch == nil {
			s.Numeric(403, target, "No such channel")
			return
		}
		ch.Mutex.RLock()
		for u := range ch.Users {
			sessions = append(sessions, u)
		}
		ch.Mutex.RUnlock()
	} else {
		// WHO nick
		s.Srv.UsersMu.RLock()
		if u, ok := s.Srv.Users[strings.ToLower(target)]; ok {
			sessions = append(sessions, u)
		}
		s.Srv.UsersMu.RUnlock()
	}

	host := s.Srv.Cfg.Server.Hostname
	me := s.Nick

	for _, u := range sessions {
		// 352 RPL_WHOREPLY: "<me> <channel> <user> <host> <server> <nick> <flags> :<real name>"
		channel := target
		if !strings.HasPrefix(target, "#") {
			channel = u.Nick
		}
		flags := "H" // H=here (no-away); you can append G if away, * if OP, etc.

		msg := &irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "352",
			Params:   []string{me, channel, u.User, host, u.Nick, flags},
			Trailing: u.Real,
		}
		s.Send(msg)
	}

	// 315 RPL_ENDOFWHO: "<me> <target> :End of WHO list"
	s.Send(&irc.Message{
		Prefix:   &irc.Prefix{Host: host},
		Command:  "315",
		Params:   []string{me, target},
		Trailing: "End of WHO list",
	})
}
