// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("NICK", Nick)
}

func Nick(s *core.Session, m *irc.Message) {
	nick := ""
	if len(m.Params) > 0 {
		nick = m.Params[0]
	} else if m.Trailing != "" {
		nick = m.Trailing
	}
	if nick == "" {
		s.Numeric(431, "No nickname given") // ERR_NONICKNAMEGIVEN
		return
	}
	s.Nick = nick
	s.TryRegister()
}
