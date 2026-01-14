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
	"fmt"
	"os"
	"time"
)

var ErrUnknownPost = errors.New("unknown post")

type Post struct {
	ID      int       `json:"id,omitempty"`
	Parent  int       `json:"parent,omitempty"`
	Name    string    `json:"name,omitempty"`
	Subject string    `json:"subject,omitempty"`
	Body    string    `json:"body,omitempty"`
	Image   bool      `json:"image,omitempty"`
	Poster  string    `json:"poster,omitempty"`
	Posted  time.Time `json:"posted,omitzero"`
	Replies PostData  `json:"replies,omitempty"`
}

func (p Post) ThumbPath() string {
	return fmt.Sprintf("data/thumb/%d.jpg", p.Posted.UnixMilli())
}

func (p Post) FullPath() string {
	return fmt.Sprintf("data/full/%d.png", p.Posted.UnixMilli())
}

func (p Post) DeleteImage() error {
	err := os.Remove(p.FullPath())
	if err != nil {
		return fmt.Errorf("failed to delete full image: %s", err)
	}

	err = os.Remove(p.ThumbPath())
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail image: %s", err)
	}

	return nil
}

type PostData []Post

type PostDB interface {
	GetAll() (PostData, error)
	Get(id int) (Post, error)
	Add(post Post) (int, error)
	Delete(id int) error
	DeletePoster(id string) error
}
