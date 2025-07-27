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
	"errors"
	"fmt"
	"os"
	"time"
)

const postersFile = "data/posters.json"

var ErrUnknownPoster = errors.New("unknown poster")

type Poster struct {
	LastPost  time.Time `json:"lastPost,omitzero"`
	LastAdmin time.Time `json:"lastAdmin,omitzero"`
	Banned    bool      `json:"banned,omitempty"`
}

type PosterData map[string]Poster

func readPosters() (PosterData, error) {
	f, err := os.Open(postersFile)
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

func writePosters(posters PosterData) error {
	f, err := os.OpenFile(postersFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

func GetPoster(ip string) (Poster, error) {
	posters, err := readPosters()
	if err != nil {
		return Poster{}, fmt.Errorf("failed to fetch posters: %s", err)
	}

	poster, ok := posters[ip]
	if !ok {
		return Poster{}, ErrUnknownPoster
	}

	return poster, nil
}

func AddPoster(ip string, poster Poster) error {
	posters, err := readPosters()
	if err != nil {
		return fmt.Errorf("failed to fetch posters: %s", err)
	}

	posters[ip] = poster

	err = writePosters(posters)
	if err != nil {
		return fmt.Errorf("failed to insert poster: %s", err)
	}

	return nil
}
