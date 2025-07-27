// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("USER", User)
}

func User(s *core.Session, m *irc.Message) {
	// USER <username> <mode> * :<realname>
	if len(m.Params) < 3 && m.Trailing == "" {
		// ERR_NEEDMOREPARAMS
		s.Numeric(461, "USER", "Not enough parameters")
		return
	}
	s.User = m.Params[0]
	if m.Trailing != "" {
		s.Real = m.Trailing
	}
	s.TryRegister()
}
