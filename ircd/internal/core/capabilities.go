// SPDX-License-Identifier: BSD-2-Clause
package core

import "sync"

var (
	capsMu  sync.RWMutex
	capList []string
	capSet  = make(map[string]struct{})
)

// RegisterCap makes a capability show up in CAP LS.
func RegisterCap(name string) {
	capsMu.Lock()
	defer capsMu.Unlock()
	if _, ok := capSet[name]; ok {
		return
	}
	capSet[name] = struct{}{}
	capList = append(capList, name)
}

// Caps returns the list of registered capabilities.
func Caps() []string {
	capsMu.RLock()
	defer capsMu.RUnlock()
	out := make([]string, len(capList))
	copy(out, capList)
	return out
}
