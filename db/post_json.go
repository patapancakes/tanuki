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
	"slices"

	. "github.com/patapancakes/tanuki/config"
)

type PostJSON struct {
	file string
}

func NewPostJSON(file string) PostJSON {
	return PostJSON{file: file}
}

func (p PostJSON) read() (PostData, error) {
	f, err := os.Open(p.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to open log file: %s", err)
	}

	defer f.Close()

	var posts PostData
	err = json.NewDecoder(f).Decode(&posts)
	if err != nil {
		return nil, fmt.Errorf("failed to decode log file: %s", err)
	}

	return posts, nil
}

func (p PostJSON) write(posts PostData) error {
	f, err := os.OpenFile(p.file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s", err)
	}

	defer f.Close()

	err = json.NewEncoder(f).Encode(posts)
	if err != nil {
		return fmt.Errorf("failed to encode log file: %s", err)
	}

	return nil
}

func (p PostJSON) GetAll() (PostData, error) {
	posts, err := p.read()
	if err != nil {
		return nil, err
	}

	// sort threads by newest reply
	slices.SortFunc(posts, func(a, b Post) int {
		t1 := a.Posted
		if len(a.Replies) != 0 {
			t1 = a.Replies[min(Config.MaxBumps, len(a.Replies))-1].Posted
		}

		t2 := b.Posted
		if len(b.Replies) != 0 {
			t2 = b.Replies[min(Config.MaxBumps, len(b.Replies))-1].Posted
		}

		return t2.Compare(t1)
	})

	return posts, nil
}

func (p PostJSON) Get(id int) (Post, error) {
	posts, err := p.read()
	if err != nil {
		return Post{}, fmt.Errorf("failed to fetch posts: %s", err)
	}

	for _, p := range posts {
		if p.ID == id {
			return p, nil
		}

		for _, reply := range p.Replies {
			if reply.ID == id {
				return reply, nil
			}
		}
	}

	return Post{}, ErrUnknownPost
}

func (p PostJSON) Add(post Post) (int, error) {
	posts, err := p.read()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch posts: %s", err)
	}

	for _, thread := range posts {
		post.ID = max(post.ID, thread.ID)

		for _, reply := range thread.Replies {
			post.ID = max(post.ID, reply.ID)
		}
	}

	post.ID++

	if post.Parent == 0 { // new thread
		posts = append(posts, post)
	} else { // new reply
		var found bool
		for i, p := range posts {
			if p.ID != post.Parent {
				continue
			}

			found = true

			posts[i].Replies = append(p.Replies, post)
			break
		}
		if !found {
			return 0, ErrUnknownPost
		}
	}

	err = p.write(posts)
	if err != nil {
		return 0, fmt.Errorf("failed to write posts: %s", err)
	}

	return post.ID, nil
}

func (p PostJSON) Delete(id int) error {
	posts, err := p.read()
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	var post Post
	for i, thread := range posts {
		for i, reply := range thread.Replies {
			if reply.ID != id {
				continue
			}

			post = reply

			posts[i].Replies = slices.Delete(thread.Replies, i, i+1)
			break
		}

		if thread.ID != id {
			continue
		}

		for _, reply := range thread.Replies {
			if !reply.Image {
				continue
			}

			err = reply.DeleteImage()
			if err != nil {
				return fmt.Errorf("failed to delete reply images: %s", err)
			}
		}

		post = thread

		posts = slices.Delete(posts, i, i+1)
		break
	}
	if post.ID == 0 {
		return ErrUnknownPost
	}
	if post.Image {
		err = post.DeleteImage()
		if err != nil {
			return fmt.Errorf("failed to delete post images: %s", err)
		}
	}

	err = p.write(posts)
	if err != nil {
		return fmt.Errorf("failed to write posts: %s", err)
	}

	return nil
}

func (p PostJSON) DeletePoster(id string) error {
	posts, err := p.read()
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	for _, thread := range posts {
		if thread.Poster == id {
			err = p.Delete(thread.ID)
			if err != nil {
				return fmt.Errorf("failed to delete post: %s", err)
			}

			continue
		}

		for _, reply := range thread.Replies {
			if reply.Poster != id {
				continue
			}

			err = p.Delete(reply.ID)
			if err != nil {
				return fmt.Errorf("failed to delete post: %s", err)
			}
		}
	}

	return nil
}
