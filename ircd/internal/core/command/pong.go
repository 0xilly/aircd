// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"time"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("PART", Part)
}

func Pong(s *core.Session, _ *irc.Message) {
	s.LastPong.Store(time.Now().UnixNano())
}
