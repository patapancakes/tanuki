/*
	tanuki - a lightweight image bbs
	Copyright (C) 2025  Pancakes (pancakes@mooglepowered.com)

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package ratelimit

import "sync"

type Mutex struct {
	entries map[string]*sync.Mutex
	mtx     sync.Mutex
}

func NewMutex() *Mutex {
	return &Mutex{
		entries: make(map[string]*sync.Mutex),
	}
}

func (m *Mutex) Lock(key string) {
	m.mtx.Lock()
	entry, exists := m.entries[key]
	if !exists {
		entry = &sync.Mutex{}
		m.entries[key] = entry
	}
	m.mtx.Unlock()

	entry.Lock()
}

func (m *Mutex) Unlock(key string) {
	m.mtx.Lock()
	entry, exists := m.entries[key]
	if !exists {
		m.mtx.Unlock()
		return
	}
	delete(m.entries, key)
	m.mtx.Unlock()

	entry.Unlock()
}
