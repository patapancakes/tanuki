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
	"strconv"
	"time"

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"
)

var adminT *template.Template

func Admin(w http.ResponseWriter, r *http.Request) {
	if Config.AdminPassword == "" {
		writeError(w, r, "admin password not set", http.StatusForbidden)
		return
	}

	err := adminT.Execute(w, nil)
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
	poster, err := GetPoster(deriveIdentity(r))
	if err != nil && err != ErrUnknownPoster {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}
	if poster.Banned {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if poster.LastAdmin.Add(time.Second * time.Duration(Config.AdminCooldown)).After(time.Now().UTC()) {
		writeError(w, r, "you are being rate limited", http.StatusTooManyRequests)
		return
	}

	poster.LastAdmin = time.Now().UTC()

	err = AddPoster(deriveIdentity(r), poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	// check password
	if r.FormValue("password") != Config.AdminPassword {
		writeError(w, r, "incorrect password", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "adminpw",
		Value:    r.FormValue("password"),
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 day
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "adminpw",
		Path:     "/",
		MaxAge:   -1, // invalidate
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AdminDelete(w http.ResponseWriter, r *http.Request) {
	if Config.AdminPassword == "" {
		writeError(w, r, "admin password not set", http.StatusForbidden)
		return
	}

	// rate limiting
	poster, err := GetPoster(deriveIdentity(r))
	if err != nil && err != ErrUnknownPoster {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}
	if poster.Banned {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if poster.LastAdmin.Add(time.Second * time.Duration(Config.AdminCooldown)).After(time.Now().UTC()) {
		writeError(w, r, "you are being rate limited", http.StatusTooManyRequests)
		return
	}

	poster.LastAdmin = time.Now().UTC()

	err = AddPoster(deriveIdentity(r), poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	// check password
	adminpw, err := r.Cookie("adminpw")
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to read admin password cookie: %s", err), http.StatusBadRequest)
		return
	}
	if adminpw.Value != Config.AdminPassword {
		writeError(w, r, "incorrect password", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to parse id: %s", err), http.StatusBadRequest)
		return
	}

	err = DeletePost(id)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to delete post: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AdminBan(w http.ResponseWriter, r *http.Request) {
	if Config.AdminPassword == "" {
		writeError(w, r, "admin password not set", http.StatusForbidden)
		return
	}

	// rate limiting
	poster, err := GetPoster(deriveIdentity(r))
	if err != nil && err != ErrUnknownPoster {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}
	if poster.Banned {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if poster.LastAdmin.Add(time.Second * time.Duration(Config.AdminCooldown)).After(time.Now().UTC()) {
		writeError(w, r, "you are being rate limited", http.StatusTooManyRequests)
		return
	}

	poster.LastAdmin = time.Now().UTC()

	err = AddPoster(deriveIdentity(r), poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	// check password
	adminpw, err := r.Cookie("adminpw")
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to read admin password cookie: %s", err), http.StatusBadRequest)
		return
	}
	if adminpw.Value != Config.AdminPassword {
		writeError(w, r, "incorrect password", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to parse id: %s", err), http.StatusBadRequest)
		return
	}

	post, err := GetPost(id)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to fetch post: %s", err), http.StatusInternalServerError)
		return
	}

	poster, err = GetPoster(post.Poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to look up poster info: %s", err), http.StatusInternalServerError)
		return
	}

	poster.Banned = true

	err = AddPoster(post.Poster, poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to insert poster: %s", err), http.StatusInternalServerError)
		return
	}

	err = DeletePosterPosts(post.Poster)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to delete poster posts: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
