// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("PASS", Pass)
}

func Pass(s *core.Session, m *irc.Message) {
	// Ignore if already registered
	if s.IsRegistered() {
		return
	}

	cfg := s.Srv.Cfg.Server
	if !cfg.PassOn {
		// PASS not required → ignore
		return
	}

	// Extract password argument
	pwd := ""
	if len(m.Params) > 0 {
		pwd = m.Params[0]
	} else if m.Trailing != "" {
		pwd = m.Trailing
	}

	if pwd == cfg.Password {
		s.PassOk = true
	} else {
		// 464 ERR_PASSWDMISMATCH
		msg := &irc.Message{
			Prefix:   &irc.Prefix{Host: s.Srv.Cfg.Server.Hostname},
			Command:  "464",
			Params:   []string{s.NickOrAsterisk()},
			Trailing: "Password incorrect",
		}
		s.Send(msg)
		s.Shutdown("bad password")
	}
}
