// SPDX-License-Identifier: BSD-2-Clause
package irc

import "fmt"

const (
	RPL_WELCOME     = 001
	RPL_MOTDSTART   = 375
	RPL_MOTD        = 372
	RPL_ENDOFMOTD   = 376
	ERR_NONICK      = 431
	ERR_NICKINUSE   = 433
	ERR_NOTREGISTER = 451
	ERR_UNKNOWNCMD  = 421
	ERR_NEEDMORE    = 461
	ERR_NOSUCHNICK  = 401
	ERR_NOSUCHCHAN  = 403
	ERR_NOTEXT      = 412
)

func Num(host string, code int, target string, args ...string) *Message {
	return &Message{
		Prefix:  &Prefix{Host: host},
		Command: fmt.Sprintf("%03d", code),
		Params:  append([]string{target}, args...),
	}
}
