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
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"net/netip"
	"os"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/patapancakes/tanuki/config"
	"github.com/patapancakes/tanuki/db"

	"github.com/xeonx/timeago"
)

var (
	funcs = template.FuncMap{
		"timeago": timeago.English.Format,
		"sum":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"max":     func(a, b int) int { return max(a, b) },
		"config":  func() ConfigFile { return Config },
		"rand":    rand.IntN,
	}

	posts   db.PostDB
	posters db.PosterDB

	//go:embed templates
	templates      embed.FS
	TemplatesFS, _ = fs.Sub(templates, "templates")

	//go:embed assets
	assets      embed.FS
	AssetsFS, _ = fs.Sub(assets, "assets")

	errInvalidSession        = errors.New("invalid session")
	errInvalidSessionSubject = errors.New("invalid session subject")
)

func Init() error {
	var err error

	// home
	homeT, err = template.New("home.html").Funcs(funcs).ParseFS(TemplatesFS, "home.html")
	if err != nil {
		return err
	}

	homeT, err = homeT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// thread
	threadT, err = template.New("thread.html").Funcs(funcs).ParseFS(TemplatesFS, "thread.html")
	if err != nil {
		return err
	}

	threadT, err = threadT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// error
	errorT, err = template.New("error.html").Funcs(funcs).ParseFS(TemplatesFS, "error.html")
	if err != nil {
		return err
	}

	errorT, err = errorT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// login
	loginT, err = template.New("login.html").Funcs(funcs).ParseFS(TemplatesFS, "login.html")
	if err != nil {
		return err
	}

	loginT, err = loginT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// bans
	bansT, err = template.New("bans.html").Funcs(funcs).ParseFS(TemplatesFS, "bans.html")
	if err != nil {
		return err
	}

	bansT, err = bansT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// confirm
	confirmT, err = template.New("confirm.html").Funcs(funcs).ParseFS(TemplatesFS, "confirm.html")
	if err != nil {
		return err
	}

	confirmT, err = confirmT.ParseFS(TemplatesFS, "include/*.html")
	if err != nil {
		return err
	}

	// database
	posts = db.NewPostJSON("data/posts.json")
	posters = db.NewPosterJSON("data/posters.json")

	return nil
}

func deriveIdentity(r *http.Request) (string, error) {
	// get ip
	addrport, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	ip := addrport.Addr()
	if addrport.Addr().IsLoopback() && r.Header.Get("X-Forwarded-For") != "" {
		ip, err = netip.ParseAddr(r.Header.Get("X-Forwarded-For"))
		if err != nil {
			return "", err
		}
	}

	return ip.String(), nil
}

func checkAuth(r *http.Request) error {
	session, err := r.Cookie("session")
	if err != nil {
		return err
	}

	token, err := jwt.Parse(session.Value, func(token *jwt.Token) (any, error) {
		return os.ReadFile("data/session.key")
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errInvalidSession
	}

	identity, err := deriveIdentity(r)
	if err != nil {
		return err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return err
	}
	if subject != identity {
		return errInvalidSessionSubject
	}

	return nil
}
