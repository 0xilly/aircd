// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	// Auto‑register the CAP handler
	core.RegisterHandler("CAP", Cap)

	// If you want to expose your capabilities here, register them too:
	core.RegisterCap("multi-prefix")
}

func Cap(s *core.Session, m *irc.Message) {
	host := s.Srv.Cfg.Server.Hostname

	sub := ""
	if len(m.Params) > 0 {
		sub = strings.ToUpper(m.Params[0])
	}

	switch sub {
	case "LS":
		caps := core.Caps()
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "CAP",
			Params:   []string{"*", "LS"},
			Trailing: strings.Join(caps, " "),
		})

	case "REQ":
		want := []string{}
		avail := make(map[string]struct{}, len(core.Caps()))
		for _, c := range core.Caps() {
			avail[c] = struct{}{}
		}
		for _, c := range strings.Split(m.Trailing, " ") {
			if _, ok := avail[c]; ok {
				want = append(want, c)
			}
		}
		cmd := "NAK"
		if len(want) > 0 {
			cmd = "ACK"
		}
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "CAP",
			Params:   []string{"*", cmd},
			Trailing: strings.Join(want, " "),
		})

	case "END":
		s.Send(&irc.Message{
			Prefix:  &irc.Prefix{Host: host},
			Command: "CAP",
			Params:  []string{"*", "END"},
		})

	default:
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "CAP",
			Params:   []string{"*", "NAK"},
			Trailing: sub,
		})
	}
}
