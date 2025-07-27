// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("NAMES", Names)
}

func Names(s *core.Session, m *irc.Message) {
	if len(m.Params) == 0 && m.Trailing == "" {
		// You can list all visible channels/users; keep it simple for now.
		s.Numeric(461, "NAMES", "Not enough parameters")
		return
	}
	targets := strings.Split(firstParam(m), ",")
	host, nick := s.Srv.Cfg.Server.Hostname, s.Nick

	for _, chName := range targets {
		low := strings.ToLower(chName)

		s.Srv.ChansMu.RLock()
		ch, ok := s.Srv.Channels[low]
		s.Srv.ChansMu.RUnlock()
		if !ok {
			s.Numeric(403, chName, "No such channel") // ERR_NOSUCHCHANNEL
			continue
		}

		// Collect member nicks
		ch.Mutex.RLock()
		names := make([]string, 0, len(ch.Users))
		for u := range ch.Users {
			// TODO: add @ for ops, + for voiced
			names = append(names, u.Nick)
		}
		ch.Mutex.RUnlock()

		// 353 RPL_NAMREPLY: <nick> = <channel> :nick1 nick2 ...
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "353",
			Params:   []string{nick, "=", chName},
			Trailing: strings.Join(names, " "),
		})
		// 366 RPL_ENDOFNAMES
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "366",
			Params:   []string{nick, chName},
			Trailing: "End of NAMES list",
		})
	}
}
