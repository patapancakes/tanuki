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
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"time"

	. "github.com/patapancakes/tanuki/config"

	"golang.org/x/image/draw"
)

var ErrUnknownPost = errors.New("unknown post")

type Post struct {
	Parent  string    `json:"parent,omitempty"`
	Name    string    `json:"name,omitempty"`
	Subject string    `json:"subject,omitempty"`
	Body    string    `json:"body,omitempty"`
	Image   bool      `json:"image,omitempty"`
	Poster  string    `json:"poster,omitempty"`
	Posted  time.Time `json:"posted,omitzero"`
	Replies PostData  `json:"replies,omitempty"`
}

func (p Post) ID() string {
	if p.Posted.IsZero() {
		return ""
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(p.Posted.UnixMilli()))

	return base64.RawURLEncoding.EncodeToString(b)
}

func (p Post) IsThread() bool {
	return p.Parent == ""
}

func (p Post) IsAdmin() bool {
	return p.Poster == "admin"
}

func (p Post) ThumbPath() string {
	return fmt.Sprintf("thumb/%s.jpg", p.ID())
}

func (p Post) FullPath() string {
	return fmt.Sprintf("full/%s.png", p.ID())
}

func (p Post) DeleteImage() error {
	err := os.Remove(path.Join("data", p.FullPath()))
	if err != nil {
		return fmt.Errorf("failed to delete full image: %s", err)
	}

	err = os.Remove(path.Join("data", p.ThumbPath()))
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail image: %s", err)
	}

	return nil
}

func (p Post) WriteImage(img image.Image) error {
	// full image
	of, err := os.OpenFile(path.Join("data", p.FullPath()), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer of.Close()

	err = png.Encode(of, img)
	if err != nil {
		return err
	}

	// thumbnail image
	scale := float64(Config.ThumbnailDimensions) / float64(img.Bounds().Dx()) // assume landscape
	if img.Bounds().Dy() >= img.Bounds().Dx() {                               // it's not
		scale = float64(Config.ThumbnailDimensions) / float64(img.Bounds().Dy())
	}

	oimg := image.NewRGBA(image.Rect(0, 0, int(scale*float64(img.Bounds().Dx())), int(scale*float64(img.Bounds().Dy()))))

	draw.Draw(oimg, oimg.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.BiLinear.Scale(oimg, oimg.Bounds(), img, img.Bounds(), draw.Over, nil)

	of, err = os.OpenFile(path.Join("data", p.ThumbPath()), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer of.Close()

	err = jpeg.Encode(of, oimg, &jpeg.Options{Quality: Config.ThumbnailQuality})
	if err != nil {
		return err
	}

	return nil
}

type PostData []Post

type PostDB interface {
	GetAll() (PostData, error)
	Get(id string) (Post, error)
	Add(post Post) (string, error)
	Delete(id string) error
	DeletePoster(id string) error
}
