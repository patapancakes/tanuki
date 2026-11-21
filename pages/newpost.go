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

package pages

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"

	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

func NewPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(Config.MaxImageSize*1024*1024))

	// admin
	var admin bool
	if Config.AdminPassword != "" {
		err := checkAuth(r)
		if err != nil {
			if err == errInvalidSession {
				http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
				return
			}
			if err != http.ErrNoCookie {
				writeError(w, r, fmt.Sprintf("authentication failed: %s", err), http.StatusUnauthorized)
				return
			}
		} else {
			admin = true
		}
	}

	// poster
	identity, err := deriveIdentity(r)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to derive identity: %s", err), http.StatusInternalServerError)
		return
	}

	poster, err := posters.Get(identity)
	if err != nil && err != ErrUnknownPoster {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}
	if poster.IsBanned() {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if time.Since(poster.LastPost) < time.Duration(Config.PostCooldown)*time.Second {
		writeError(w, r, "you are posting too quickly", http.StatusTooManyRequests)
		return
	}

	// post
	var post Post

	post.Poster = identity
	if admin {
		post.Poster = "admin"
	}

	post.Name = strings.TrimSpace(r.PostFormValue("name"))
	if !utf8.ValidString(post.Name) || utf8.RuneCountInString(post.Name) > Config.MaxNameLength {
		writeError(w, r, "invalid name", http.StatusBadRequest)
		return
	}

	post.Subject = strings.TrimSpace(r.PostFormValue("subject"))
	if !utf8.ValidString(post.Subject) || utf8.RuneCountInString(post.Subject) > Config.MaxSubjectLength {
		writeError(w, r, "invalid subject", http.StatusBadRequest)
		return
	}

	post.Body = strings.TrimSpace(r.PostFormValue("comment"))
	if !utf8.ValidString(post.Body) || utf8.RuneCountInString(post.Body) > Config.MaxCommentLength {
		writeError(w, r, "invalid comment", http.StatusBadRequest)
		return
	}

	post.Parent = r.PostFormValue("parent")
	post.Posted = time.Now()

	// handle image
	f, _, err := r.FormFile("image")
	if err != nil {
		if err != http.ErrMissingFile {
			writeError(w, r, fmt.Sprintf("failed to parse form file: %s", err), http.StatusBadRequest)
			return
		}
	} else {
		cfg, format, err := image.DecodeConfig(f)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to decode image file: %s", err), http.StatusBadRequest)
			return
		}
		if cfg.Width*cfg.Height > Config.MaxImagePixels*1000*1000 {
			writeError(w, r, "image dimensions are too large", http.StatusBadRequest)
			return
		}

		post.Image = format

		f.Seek(0, io.SeekStart)

		err = post.WriteImage(f)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to write image files: %s", err), http.StatusInternalServerError)
			return
		}
	}

	if post.Body == "" && !post.HasImage() {
		writeError(w, r, "a comment or image is required", http.StatusBadRequest)
		return
	}

	err = posters.Add(identity, Poster{LastPost: post.Posted})
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	var id string
	id, err = posts.Add(post)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert post: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("%s#post_%s", post.Parent, post.ID())
	if post.IsThread() {
		redirect = id
	}

	http.Redirect(w, r, fmt.Sprintf("/thread/%s", redirect), http.StatusFound)

	postTypeText := "thread"
	if !post.IsThread() {
		postTypeText = fmt.Sprintf("reply to thread \"%s\"", post.Parent)
	}

	writeLog(r, fmt.Sprintf("created new %s with id \"%s\"", postTypeText, post.ID()))
}
