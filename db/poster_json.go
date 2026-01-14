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
	"encoding/json"
	"fmt"
	"os"
)

type PosterJSON struct {
	file string
}

func NewPosterJSON(file string) PosterJSON {
	return PosterJSON{file: file}
}

func (p PosterJSON) read() (PosterData, error) {
	f, err := os.Open(p.file)
	if err != nil {
		if os.IsNotExist(err) {
			return make(PosterData), nil
		}

		return nil, fmt.Errorf("failed to open posters file: %s", err)
	}

	defer f.Close()

	posters := make(PosterData)
	err = json.NewDecoder(f).Decode(&posters)
	if err != nil {
		return nil, fmt.Errorf("failed to decode posters file: %s", err)
	}

	return posters, nil
}

func (p PosterJSON) write(posters PosterData) error {
	f, err := os.OpenFile(p.file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open posters file: %s", err)
	}

	defer f.Close()

	err = json.NewEncoder(f).Encode(posters)
	if err != nil {
		return fmt.Errorf("failed to encode posters file: %s", err)
	}

	return nil
}

func (p PosterJSON) Get(id string) (Poster, error) {
	posters, err := p.read()
	if err != nil {
		return Poster{}, fmt.Errorf("failed to fetch posters: %s", err)
	}

	poster, ok := posters[id]
	if !ok {
		return Poster{}, ErrUnknownPoster
	}

	return poster, nil
}

func (p PosterJSON) Add(id string, poster Poster) error {
	posters, err := p.read()
	if err != nil {
		return fmt.Errorf("failed to fetch posters: %s", err)
	}

	posters[id] = poster

	err = p.write(posters)
	if err != nil {
		return fmt.Errorf("failed to insert poster: %s", err)
	}

	return nil
}
