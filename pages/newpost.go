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
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"

	_ "image/gif"

	_ "golang.org/x/image/bmp"

	"golang.org/x/image/draw"
)

func NewPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(Config.MaxUploadSize)*1024)

	// poster
	poster, err := GetPoster(deriveIdentity(r))
	if err != nil && err != ErrUnknownPoster {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}
	if poster.Banned {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if poster.LastPost.Add(time.Second * time.Duration(Config.PostCooldown)).After(time.Now().UTC()) {
		writeError(w, r, "you are being rate limited", http.StatusTooManyRequests)
		return
	}

	// post
	var post Post

	post.Poster = deriveIdentity(r)

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

	if r.PostFormValue("parent") != "" {
		post.Parent, err = strconv.Atoi(r.PostFormValue("parent"))
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to parse parent value: %s", err), http.StatusBadRequest)
			return
		}
		if post.Parent < 0 {
			writeError(w, r, "invalid parent value", http.StatusBadRequest)
			return
		}
	}

	post.Posted = time.Now().UTC()

	// handle image
	f, _, err := r.FormFile("image")
	if err != nil {
		if err != http.ErrMissingFile {
			writeError(w, r, fmt.Sprintf("failed to parse form file: %s", err), http.StatusBadRequest)
			return
		}
	} else {
		post.Image = true

		img, _, err := image.Decode(f)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to decode image file: %s", err), http.StatusBadRequest)
			return
		}

		// full image
		of, err := os.OpenFile(post.FullPath(), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to open output image file for writing: %s", err), http.StatusInternalServerError)
			return
		}

		defer of.Close()

		err = png.Encode(of, img)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to encode image: %s", err), http.StatusInternalServerError)
			return
		}

		// thumbnail image
		scale := float64(Config.ThumbnailDimensions) / float64(img.Bounds().Dx()) // assume landscape
		if img.Bounds().Dy() >= img.Bounds().Dx() {                               // it's not
			scale = float64(Config.ThumbnailDimensions) / float64(img.Bounds().Dy())
		}

		oimg := image.NewRGBA(image.Rect(0, 0, int(scale*float64(img.Bounds().Dx())), int(scale*float64(img.Bounds().Dy()))))

		draw.Draw(oimg, oimg.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.BiLinear.Scale(oimg, oimg.Bounds(), img, img.Bounds(), draw.Over, nil)

		of, err = os.OpenFile(post.ThumbPath(), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to open output image file for writing: %s", err), http.StatusInternalServerError)
			return
		}

		defer of.Close()

		err = jpeg.Encode(of, oimg, &jpeg.Options{Quality: Config.ThumbnailQuality})
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to encode image: %s", err), http.StatusInternalServerError)
			return
		}
	}

	if post.Body == "" && !post.Image {
		writeError(w, r, "a comment or image is required", http.StatusBadRequest)
		return
	}

	err = AddPoster(deriveIdentity(r), Poster{LastPost: post.Posted})
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	post.ID, err = AddPost(post)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert post: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := post.Parent
	if post.Parent == 0 {
		redirect = post.ID
	}

	http.Redirect(w, r, fmt.Sprintf("/thread/%d", redirect), http.StatusFound)
}
