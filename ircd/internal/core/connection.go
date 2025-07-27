// SPDX-License-Identifier: BSD-2-Clause
package core

import (
	"context"

	"github.com/0xilly/aircd/ircd/internal/irc"
)

type Connection interface {
	ReadLine() (string, error) // one IRC line (w/o CRLF), ≤ MaxLineBytes
	Send(*irc.Message) error   // thread-safe
	Context() context.Context  // cancels when closed
	Addr() string
	Close() error
}
