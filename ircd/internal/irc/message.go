// SPDX-License-Identifier: BSD-2-Clause
package irc

import "strings"

type Prefix struct {
	Nick string
	User string
	Host string
}

type Message struct {
	Prefix   *Prefix
	Command  string
	Params   []string
	Trailing string
}

// note(anita): might need to optimize this but meh for now
func (p *Prefix) String() string {
	if p == nil {
		return ""
	}
	// nick[!user][@host]  OR just host/servername
	if p.Nick != "" {
		var b strings.Builder
		b.WriteString(p.Nick)
		if p.User != "" {
			b.WriteByte('!')
			b.WriteString(p.User)
		}
		if p.Host != "" {
			b.WriteByte('@')
			b.WriteString(p.Host)
		}
		return b.String()
	}
	return p.Host
}

// note(anita): might need to optimize this but meh for now
// String returns the RFC1459 line WITHOUT the trailing CRLF.
func (m *Message) String() string {
	var b strings.Builder

	// Optional prefix
	if m.Prefix != nil {
		b.WriteByte(':')
		b.WriteString(m.Prefix.String())
		b.WriteByte(' ')
	}

	// Command
	b.WriteString(m.Command)

	// Middle params
	for _, p := range m.Params {
		b.WriteByte(' ')
		b.WriteString(p)
	}

	// Trailing (last) param
	if m.Trailing != "" {
		b.WriteString(" :")
		// IRC forbids CR/LF inside; strip just in case
		tr := strings.ReplaceAll(strings.ReplaceAll(m.Trailing, "\r", ""), "\n", "")
		b.WriteString(tr)
	}

	return b.String()
}
