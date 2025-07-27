// SPDX-License-Identifier: BSD-2-Clause
package core

import (
	"github.com/0xilly/aircd/ircd/internal/irc"
)

type Dispatcher struct {
	handlers map[string]Handler
	srv      *Server
}

// todo move
var preReg = map[string]struct{}{
	"NICK": {}, "USER": {}, "PASS": {}, "CAP": {}, "PING": {}, "PONG": {}, "QUIT": {},
}

func NewDispatcher(srv *Server) *Dispatcher {
	return &Dispatcher{srv: srv, handlers: make(map[string]Handler)}
}

func (d *Dispatcher) Dispatch(s *Session, m *irc.Message) {
	if !s.ready {
		if _, ok := preReg[m.Command]; !ok {
			s.Numeric(451, "You have not registered") // ERR_NOTREGISTERED
			return
		}
	}
	if h, ok := d.handlers[m.Command]; ok {
		h(s, m)
	} else {
		s.Numeric(421, m.Command, "Unknown command") // ERR_UNKNOWNCOMMAND
	}
}
