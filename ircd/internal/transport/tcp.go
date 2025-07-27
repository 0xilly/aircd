// SPDX-License-Identifier: BSD-2-Clause
package transport

import (
	"bufio"
	"context"
	"net"
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

type tcpConn struct {
	c      net.Conn
	r      *bufio.Reader
	w      *bufio.Writer
	ctx    context.Context
	cancel context.CancelFunc
	limit  int // max line bytes
}

func (t *tcpConn) Context() context.Context { return t.ctx }
func (t *tcpConn) Addr() string             { return t.c.RemoteAddr().String() }
func (t *tcpConn) Close() error             { t.cancel(); return t.c.Close() }

func (t *tcpConn) ReadLine() (string, error) {
	// Read until LF; RFC1459 line limit 512 incl CRLF, so enforce.
	line, err := t.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Trim CRLF
	line = strings.TrimRight(line, "\r\n")
	if len(line) > t.limit {
		line = line[:t.limit]
	}
	return line, nil
}

func (t *tcpConn) Send(m *irc.Message) error {
	// NOTE: no locking shown; add a mutex if multiple goroutines call Send.
	_, err := t.w.WriteString(m.String())
	if err != nil {
		return err
	}
	_, err = t.w.WriteString("\r\n")
	if err != nil {
		return err
	}
	return t.w.Flush()
}

func ListenTCP(addr string, srv *core.Server, maxLine int) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		raw, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(nc net.Conn) {
			ctx, cancel := context.WithCancel(context.Background())
			tc := &tcpConn{
				c:      nc,
				r:      bufio.NewReader(nc),
				w:      bufio.NewWriter(nc),
				ctx:    ctx,
				cancel: cancel,
				limit:  maxLine,
			}
			// Accept DOES NOT need to be in its own goroutine if it starts Session.Run itself.
			srv.Accept(tc)
		}(raw)
	}
}
