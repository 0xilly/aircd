// SPDX-License-Identifier: BSD-2-Clause
package core

import (
	"sync"

	"github.com/0xilly/aircd/ircd/internal/irc"
)

type Channel struct {
	Name  string
	Mutex sync.RWMutex
	Users map[*Session]struct{}

	topic string
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:  name,
		Users: map[*Session]struct{}{},
	}
}

func (c *Channel) Topic() string {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.topic
}

func (c *Channel) SetTopic(t string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.topic = t
}

func (c *Channel) Broadcast(from *Session, msg *irc.Message) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	for u := range c.Users {
		if u != from {
			u.Send(msg)
		}
	}
}

func (s *Server) ChannelList() []*Channel {
	s.ChansMu.RLock()
	defer s.ChansMu.RUnlock()
	out := make([]*Channel, 0, len(s.Channels))
	for _, ch := range s.Channels {
		out = append(out, ch)
	}
	return out
}

func (c *Channel) UsersCount() int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return len(c.Users)
}

func (c *Channel) Remove(u *Session) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	delete(c.Users, u)
}
