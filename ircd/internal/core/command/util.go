// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/irc"
)

func firstParam(m *irc.Message) string {
	if len(m.Params) > 0 {
		return m.Params[0]
	}
	return m.Trailing
}

func isChannel(name string) bool {
	// starts with any of the configured prefixes
	if name == "" {
		return false
	}
	p := name[0]
	return strings.ContainsRune("#", rune(p)) // or s.srv.Cfg.Channels.Prefixes
}
