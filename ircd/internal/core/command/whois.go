// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("WHOIS", Whois)
}

// WHOIS <nick>
func Whois(s *core.Session, m *irc.Message) {
	if len(m.Params) < 1 {
		s.Numeric(461, "WHOIS", "Not enough parameters")
		return
	}
	target := m.Params[0]
	host := s.Srv.Cfg.Server.Hostname
	me := s.Nick

	// Find the user session
	s.Srv.UsersMu.RLock()
	u := s.Srv.Users[strings.ToLower(target)]
	s.Srv.UsersMu.RUnlock()
	if u == nil {
		s.Numeric(401, target, "No such nick") // ERR_NOSUCHNICK
		return
	}

	// 311 RPL_WHOISUSER: me target user host server nick :realname
	s.Send(&irc.Message{
		Prefix:   &irc.Prefix{Host: host},
		Command:  "311",
		Params:   []string{me, target, u.User, host, host, target},
		Trailing: u.Real,
	})

	// 312 RPL_WHOISSERVER: me target server :server info
	s.Send(&irc.Message{
		Prefix:   &irc.Prefix{Host: host},
		Command:  "312",
		Params:   []string{me, target, host},
		Trailing: "AirCD IRC Server",
	})

	// 319 RPL_WHOISCHANNELS: me target :#chan1 #chan2 ...
	var chans []string
	s.Srv.ChansMu.RLock()
	for _, ch := range s.Srv.Channels {
		ch.Mutex.RLock()
		if _, ok := ch.Users[u]; ok {
			chans = append(chans, ch.Name)
		}
		ch.Mutex.RUnlock()
	}
	s.Srv.ChansMu.RUnlock()
	if len(chans) > 0 {
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "319",
			Params:   []string{me, target},
			Trailing: strings.Join(chans, " "),
		})
	}

	// 318 RPL_ENDOFWHOIS: me target :End of WHOIS list
	s.Send(&irc.Message{
		Prefix:   &irc.Prefix{Host: host},
		Command:  "318",
		Params:   []string{me, target},
		Trailing: "End of WHOIS list",
	})
}
