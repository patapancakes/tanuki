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

package db

import (
	"errors"
	"time"
)

var ErrUnknownPoster = errors.New("unknown poster")

type Poster struct {
	LastPost  time.Time `json:"lastPost,omitzero"`
	LastAdmin time.Time `json:"lastAdmin,omitzero"`
	Banned    bool      `json:"banned,omitempty"`
}

type PosterData map[string]Poster

type PosterDB interface {
	Get(id string) (Poster, error)
	Add(id string, poster Poster) error
}
