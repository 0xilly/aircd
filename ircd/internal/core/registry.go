// SPDX-License-Identifier: BSD-2-Clause
package core

import "github.com/0xilly/aircd/ircd/internal/irc"

type Handler func(*Session, *irc.Message)

var defaultHandlers = make(map[string]Handler)

func RegisterHandler(cmd string, h Handler) {
	defaultHandlers[cmd] = h
}
