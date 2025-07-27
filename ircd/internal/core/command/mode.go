// SPDX-License-Identifier: BSD-2-Clause
package command

import (
	"strings"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
)

func init() {
	core.RegisterHandler("MODE", Mode)
}

// Mode implements the IRC MODE command for both users and channels.
func Mode(s *core.Session, m *irc.Message) {

	// Need at least a target (nick or channel)
	if len(m.Params) < 1 {
		s.Numeric(461, "MODE", "Not enough parameters")
		return
	}
	target := m.Params[0]

	// === 1) User mode ===
	if strings.EqualFold(target, s.Nick) {
		// No mode‑change args? Show current user‑mode
		s.Numeric(221, s.Nick, "+i") // stub: show +i always
		return
	}

	// === 2) Channel mode ===
	// Must have channel prefix (e.g. #)
	if !strings.HasPrefix(target, "#") {
		s.Numeric(472, target, "is not a channel") // ERR_UNKNOWNMODE
		return
	}

	// Look up channel
	ch := s.Srv.GetChannel(target)
	if ch == nil {
		s.Numeric(403, target, "No such channel")
		return
	}

	// If only one param, show current channel modes
	if len(m.Params) < 2 {
		s.Numeric(324, s.Nick, target, "+nt") // fixme: stub: show +nt
		return
	}

	// === 3) Mode change request ===
	modeChange := m.Params[1] // e.g. "+o nick" or "-t"
	args := []string{target, modeChange}
	// any further args (e.g. nick to op) get appended
	if len(m.Params) > 2 {
		args = append(args, m.Params[2:]...)
	}

	// Broadcast the MODE change to the channel
	msg := &irc.Message{
		Prefix:  s.Prefix(),
		Command: "MODE",
		Params:  args,
	}
	// Send to all members (including changer)
	ch.Broadcast(s, msg)
}
