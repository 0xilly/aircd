// SPDX-License-Identifier: BSD-2-Clause
package core

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xilly/aircd/ircd/internal/irc"
)

type Session struct {
	Srv    *Server
	Conn   Connection
	Nick   string
	User   string
	Real   string
	ready  bool
	PassOk bool

	out chan *irc.Message

	closeOnce  sync.Once
	regTimer   *time.Timer
	LastPong   atomic.Int64
	pingTicker *time.Ticker
}

func NewSession(s *Server, c Connection) *Session {
	sess := &Session{
		Srv:  s,
		Conn: c,
		out:  make(chan *irc.Message, s.Cfg.Limits.WriteQueue),
	}
	// Kill connection if not registered in time
	if to := s.Cfg.Limits.RegistrationTimeout; to > 0 {
		sess.regTimer = time.AfterFunc(to, func() {
			if !sess.ready {
				sess.Shutdown("registration timeout")
			}
		})
	}
	return sess
}

func (s *Session) Run() {
	// writer goroutine
	go s.writePump()

	for {
		select {
		case <-s.Conn.Context().Done():
			s.Shutdown("connection closed")
			return
		default:
		}

		line, err := s.Conn.ReadLine()
		if err != nil {
			s.Shutdown("read error")
			return
		}
		msg, perr := irc.ParseLine(line)
		if perr != nil {
			// optionally Send ERR_UNKNOWNCOMMAND etc.
			continue
		}
		s.Srv.dispatcher.Dispatch(s, msg)
	}
}

func (s *Session) writePump() {
	for msg := range s.out {
		if err := s.Conn.Send(msg); err != nil {
			s.Shutdown("write error")
			return
		}
	}
}

func (s *Session) Send(m *irc.Message) {
	select {
	case s.out <- m:
	default:
		// buffer full -> drop or close; here we close:
		s.Shutdown("Send buffer overflow")
	}
}

func (s *Session) TryRegister() {
	if s.ready || s.Nick == "" || s.User == "" {
		return
	}

	//PassCheck
	if s.Srv.Cfg.Server.PassOn && !s.PassOk {
		// 464 ERR_PASSWDMISMATCH
		s.Numeric(464, "*", "Password incorrect")
		return
	}

	lower := strings.ToLower(s.Nick)

	// Nick collision check
	s.Srv.UsersMu.Lock()
	if _, exists := s.Srv.Users[lower]; exists {
		s.Srv.UsersMu.Unlock()
		s.Numeric(irc.ERR_NICKINUSE, s.Nick, "Nickname is already in use")
		return
	}
	s.Srv.Users[lower] = s
	s.Srv.UsersMu.Unlock()

	// We’re registered now
	s.ready = true
	if s.regTimer != nil {
		s.regTimer.Stop()
	}
	s.startHeartbeat()

	host := s.Srv.Cfg.Server.Hostname
	network := s.Srv.Cfg.Server.Network
	Nick := s.Nick

	// 001 RPL_WELCOME
	s.Numeric(irc.RPL_WELCOME, fmt.Sprintf("Welcome to %s %s!%s@%s", network, Nick, s.User, host))

	// (Optional) Send more 002/003/etc here

	// MOTD
	motd := s.Srv.getMOTD()
	if len(motd) == 0 {
		s.Numeric(422, "MOTD File is missing") // ERR_NOMOTD
	} else {
		s.Numeric(irc.RPL_MOTDSTART, ":- "+host+" Message of the Day -")
		for _, line := range motd {
			s.Numeric(irc.RPL_MOTD, ":- "+line)
		}
		s.Numeric(irc.RPL_ENDOFMOTD, ":End of MOTD command")
	}

	// Optional: start ping loop if you implemented it
	// s.startPingLoop()
}

func (s *Session) NickOrAsterisk() string {
	if s.Nick != "" {
		return s.Nick
	}
	return "*"
}

func (s *Session) IsRegistered() bool {
	return s.ready
}

func (s *Session) startHeartbeat() {
	iv := s.Srv.Cfg.Limits.PingInterval
	to := s.Srv.Cfg.Limits.PingTimeout
	if iv <= 0 || to <= 0 {
		return // disabled
	}

	s.LastPong.Store(time.Now().UnixNano())
	s.pingTicker = time.NewTicker(iv)

	go func() {
		host := s.Srv.Cfg.Server.Hostname
		for range s.pingTicker.C {
			if s.Conn.Context().Err() != nil {
				return
			}
			token := strconv.FormatInt(time.Now().UnixNano(), 10)
			s.Send(&irc.Message{
				Prefix:   &irc.Prefix{Host: host},
				Command:  "PING",
				Params:   []string{host},
				Trailing: token,
			})

			// wait timeout window
			time.Sleep(to)
			last := time.Unix(0, s.LastPong.Load())
			if time.Since(last) > to {
				s.Shutdown("ping timeout")
				return
			}
		}
	}()
}

func (s *Session) Shutdown(reason string) {
	s.closeOnce.Do(func() {
		if s.pingTicker != nil {
			s.pingTicker.Stop()
		}
		if reason == "" {
			reason = "closing"
		}

		// If registered, broadcast QUIT and clean up maps
		if s.ready {
			quit := &irc.Message{
				Prefix:   s.Prefix(),
				Command:  "QUIT",
				Trailing: reason,
			}

			// Remove from channels & notify others
			s.Srv.ChansMu.RLock()
			for _, ch := range s.Srv.Channels {
				ch.Mutex.Lock()
				if _, ok := ch.Users[s]; ok {
					delete(ch.Users, s)
					ch.Mutex.Unlock()
					ch.Broadcast(s, quit)
				} else {
					ch.Mutex.Unlock()
				}
			}
			s.Srv.ChansMu.RUnlock()

			// Remove from Users map
			s.Srv.UsersMu.Lock()
			delete(s.Srv.Users, strings.ToLower(s.Nick))
			s.Srv.UsersMu.Unlock()
		}

		close(s.out)
		_ = s.Conn.Close()
	})
}

func (s *Session) Prefix() *irc.Prefix {
	return &irc.Prefix{
		Nick: s.Nick,
		User: s.User,
		Host: s.Srv.Cfg.Server.Hostname,
	}
}

func (s *Session) Numeric(code int, args ...string) {
	// RFC: first param after code is the target Nick
	params := append([]string{s.Nick}, args...)
	msg := &irc.Message{
		Prefix:  &irc.Prefix{Host: s.Srv.Cfg.Server.Hostname},
		Command: fmt.Sprintf("%03d", code),
		Params:  params,
	}
	s.Send(msg)
}
