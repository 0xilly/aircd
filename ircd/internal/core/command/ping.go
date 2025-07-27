// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("PING", Ping)
}

func Ping(s *core.Session, m *irc.Message) {
	arg := ""
	if len(m.Params) > 0 {
		arg = m.Params[0]
	} else if m.Trailing != "" {
		arg = m.Trailing
	}
	reply := &irc.Message{
		Prefix:  &irc.Prefix{Host: s.Srv.Cfg.Server.Hostname},
		Command: "PONG",
		Params:  []string{s.Srv.Cfg.Server.Hostname},
	}
	if arg != "" {
		reply.Params = append(reply.Params, arg)
	}
	s.Send(reply)
}
