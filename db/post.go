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
	"slices"
	"time"
)

const postsFile = "data/posts.json"

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

type PostData []Post

func readPosts() (PostData, error) {
	f, err := os.Open(postsFile)
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

func writePosts(posts PostData) error {
	f, err := os.OpenFile(postsFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

func GetPosts() (PostData, error) {
	return readPosts()
}

func GetPost(id int) (Post, error) {
	posts, err := readPosts()
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

func AddPost(post Post) (int, error) {
	posts, err := readPosts()
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

	err = writePosts(posts)
	if err != nil {
		return 0, fmt.Errorf("failed to write posts: %s", err)
	}

	return post.ID, nil
}

func DeletePost(id int) error {
	posts, err := readPosts()
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

			err = DeletePostImages(reply)
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
		err = DeletePostImages(post)
		if err != nil {
			return fmt.Errorf("failed to delete post images: %s", err)
		}
	}

	err = writePosts(posts)
	if err != nil {
		return fmt.Errorf("failed to write posts: %s", err)
	}

	return nil
}

func DeletePostImages(post Post) error {
	err := os.Remove(post.FullPath())
	if err != nil {
		return fmt.Errorf("failed to delete full image: %s", err)
	}

	err = os.Remove(post.ThumbPath())
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail image: %s", err)
	}

	return nil
}

func DeletePosterPosts(id string) error {
	posts, err := readPosts()
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	for _, thread := range posts {
		if thread.Poster == id {
			err = DeletePost(thread.ID)
			if err != nil {
				return fmt.Errorf("failed to delete post: %s", err)
			}

			continue
		}

		for _, reply := range thread.Replies {
			if reply.Poster != id {
				continue
			}

			err = DeletePost(reply.ID)
			if err != nil {
				return fmt.Errorf("failed to delete post: %s", err)
			}
		}
	}

	return nil
}
