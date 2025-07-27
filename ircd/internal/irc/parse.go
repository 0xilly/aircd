// SPDX-License-Identifier: BSD-2-Clause
package irc

import (
	"errors"
	"strings"
)

var (
	ErrEmptyLine = errors.New("irc: empty line")
	ErrNoCommand = errors.New("irc: missing command")
)

const maxParams = 15 // RFC2812: 14 middle + 1 trailing

// ParseLine parses a single IRC line (WITHOUT the trailing CRLF) into a Message.
func ParseLine(line string) (*Message, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, ErrEmptyLine
	}

	var prefixStr string

	// Parse optional prefix
	if line[0] == ':' {
		var rest string
		prefixStr, rest, _ = strings.Cut(line[1:], " ")
		if rest == "" {
			return nil, ErrNoCommand
		}
		line = strings.TrimLeft(rest, " ")
	}

	if line == "" {
		return nil, ErrNoCommand
	}

	// Parse command
	cmd, rest, found := strings.Cut(line, " ")
	if !found {
		rest = ""
	}
	cmd = strings.ToUpper(cmd)
	if cmd == "" {
		return nil, ErrNoCommand
	}
	line = strings.TrimLeft(rest, " ")

	// Parse parameters and trailing
	params := make([]string, 0, maxParams-1)
	var trailing string

	for len(params) < maxParams-1 && line != "" {
		if line[0] == ':' {
			trailing = line[1:]
			break
		}
		param, rest, found := strings.Cut(line, " ")
		params = append(params, param)
		if !found {
			break
		}
		line = strings.TrimLeft(rest, " ")
	}

	// Build message
	var pref *Prefix
	if prefixStr != "" {
		pref = parsePrefix(prefixStr)
	}

	return &Message{
		Prefix:   pref,
		Command:  cmd,
		Params:   params,
		Trailing: trailing,
	}, nil
}

func parsePrefix(s string) *Prefix {
	p := &Prefix{}
	// nick!user@host  |  nick@host  |  nick!user  |  host
	if bang := strings.IndexByte(s, '!'); bang != -1 {
		p.Nick = s[:bang]
		rest := s[bang+1:]
		if at := strings.IndexByte(rest, '@'); at != -1 {
			p.User = rest[:at]
			p.Host = rest[at+1:]
		} else {
			p.User = rest
		}
		return p
	}
	if at := strings.IndexByte(s, '@'); at != -1 {
		p.Nick = s[:at]
		p.Host = s[at+1:]
		return p
	}
	// servername / host only
	p.Host = s
	return p
}
