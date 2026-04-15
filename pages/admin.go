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
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"
)

var loginT *template.Template

func Login(w http.ResponseWriter, r *http.Request) {
	if Config.AdminPassword == "" {
		writeError(w, r, "admin password not set", http.StatusForbidden)
		return
	}

	err := loginT.Execute(w, r.Header.Get("Referer"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to execute template: %s", err), http.StatusInternalServerError)
		return
	}
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	if Config.AdminPassword == "" {
		writeError(w, r, "admin password not set", http.StatusForbidden)
		return
	}

	// rate limiting
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
	if poster.LastLogin.Add(time.Second * time.Duration(Config.AdminCooldown)).After(time.Now()) {
		writeError(w, r, "you are logging in too quickly", http.StatusTooManyRequests)
		return
	}

	poster.LastLogin = time.Now()

	err = posters.Add(identity, poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	// check password
	if r.FormValue("password") != Config.AdminPassword {
		writeError(w, r, "incorrect password", http.StatusUnauthorized)
		return
	}

	key, err := os.ReadFile("data/session.key")
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to read session signing key: %s", err), http.StatusInternalServerError)
		return
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   identity,
	}).SignedString(key)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to sign token: %s", err), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 day
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	redirect := r.FormValue("referer")
	if redirect == "" {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Path:     "/",
		MaxAge:   -1, // invalidate
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	redirect := r.Header.Get("Referer")
	if redirect == "" {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func AdminDelete(w http.ResponseWriter, r *http.Request) {
	err := checkAuth(r)
	if err != nil {
		writeError(w, r, fmt.Sprintf("authentication failed: %s", err), http.StatusUnauthorized)
		return
	}

	err = posts.Delete(r.FormValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to delete post: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := r.FormValue("referer")
	if redirect == "" {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func AdminBan(w http.ResponseWriter, r *http.Request) {
	err := checkAuth(r)
	if err != nil {
		writeError(w, r, fmt.Sprintf("authentication failed: %s", err), http.StatusUnauthorized)
		return
	}

	post, err := posts.Get(r.FormValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to fetch post: %s", err), http.StatusInternalServerError)
		return
	}

	poster, err := posters.Get(post.Poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}

	poster.BanTime = time.Now()
	poster.BanReason = r.FormValue("reason")

	err = posters.Add(post.Poster, poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	err = posts.DeletePoster(post.Poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to delete poster posts: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := r.FormValue("referer")
	if redirect == "" {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func AdminUnbanID(w http.ResponseWriter, r *http.Request) {
	err := checkAuth(r)
	if err != nil {
		writeError(w, r, fmt.Sprintf("authentication failed: %s", err), http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	ids, ok := r.Form["id"]
	if !ok {
		writeError(w, r, "no ids specified", http.StatusBadRequest)
		return
	}

	for _, id := range ids {
		poster, err := posters.Get(id)
		if err != nil && err != ErrUnknownPoster {
			writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
			return
		}

		poster.BanTime = time.Time{}

		err = posters.Add(id, poster)
		if err != nil {
			writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
			return
		}
	}

	redirect := r.Header.Get("Referer")
	if redirect == "" {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
