// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strconv"
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("LIST", List)
}

func List(s *core.Session, m *irc.Message) {
	host := s.Srv.Cfg.Server.Hostname
	me := s.Nick

	// If no params, list every channel; else only those named
	var toList []string
	if len(m.Params) == 0 {
		for _, ch := range s.Srv.ChannelList() {
			toList = append(toList, ch.Name)
		}
	} else {
		toList = strings.Split(m.Params[0], ",")
	}

	for _, name := range toList {
		ch := s.Srv.GetChannel(name)
		if ch == nil {
			// silently skip non‑existent channels
			continue
		}
		// 322 RPL_LIST: <me> <channel> <# visible> :<topic>
		s.Send(&irc.Message{
			Prefix:   &irc.Prefix{Host: host},
			Command:  "322",
			Params:   []string{me, ch.Name, strconv.Itoa(ch.UsersCount())},
			Trailing: ch.Topic(),
		})
	}

	// 323 RPL_LISTEND: <me> :End of LIST
	s.Send(&irc.Message{
		Prefix:   &irc.Prefix{Host: host},
		Command:  "323",
		Params:   []string{me},
		Trailing: "End of LIST",
	})
}
