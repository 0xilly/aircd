// SPDX-License-Identifier: BSD-2-Clause
package core

import (
	"os"
	"strings"
	"sync"

	cfg "github.com/0xilly/aircd/ircd/internal/config"
)

type Server struct {
	Cfg        *cfg.Config
	dispatcher *Dispatcher

	MotdMu sync.RWMutex
	Motd   []string

	UsersMu sync.RWMutex
	Users   map[string]*Session

	ChansMu  sync.RWMutex
	Channels map[string]*Channel
}

func NewServer(c *cfg.Config) *Server {
	s := &Server{
		Cfg:      c,
		Users:    make(map[string]*Session),
		Channels: make(map[string]*Channel),
	}
	s.dispatcher = NewDispatcher(s)
	return s
}

func (s *Server) Accept(conn Connection) {
	sess := NewSession(s, conn)
	go sess.Run()
}

func (s *Server) GetOrCreateChannel(name string) *Channel {
	low := strings.ToLower(name)
	s.ChansMu.RLock()
	ch, ok := s.Channels[low]
	s.ChansMu.RUnlock()
	if ok {
		return ch
	}
	s.ChansMu.Lock()
	defer s.ChansMu.Unlock()
	if ch, ok := s.Channels[low]; ok {
		return ch
	}
	ch = NewChannel(name)
	s.Channels[low] = ch
	return ch
}

func (s *Server) GetChannel(name string) *Channel {
	low := strings.ToLower(name)
	s.ChansMu.RLock()
	ch := s.Channels[low]
	s.ChansMu.RUnlock()
	return ch
}

func (s *Server) getMOTD() []string {
	s.MotdMu.RLock()
	defer s.MotdMu.RUnlock()
	// copy to avoid races if caller iterates
	out := make([]string, len(s.Motd))
	copy(out, s.Motd)
	return out
}

func (s *Server) ReloadMOTD() error {
	lines := loadMOTD(s.Cfg.Server.MOTDFile)

	s.MotdMu.Lock()
	s.Motd = lines
	s.MotdMu.Unlock()
	return nil
}

func loadMOTD(path string) []string {
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	txt := strings.ReplaceAll(string(b), "\r\n", "\n")
	lines := strings.Split(txt, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
