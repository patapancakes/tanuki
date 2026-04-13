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
	"time"

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"
)

var bansT *template.Template

func Bans(w http.ResponseWriter, r *http.Request) {
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
	if poster.Banned {
		writeError(w, r, "you are banned", http.StatusForbidden)
		return
	}
	if poster.LastAdmin.Add(time.Second * time.Duration(Config.AdminCooldown)).After(time.Now()) {
		writeError(w, r, "you are being rate limited", http.StatusTooManyRequests)
		return
	}

	poster.LastAdmin = time.Now()

	err = posters.Add(identity, poster)
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

	banned, err := posters.GetBanned()
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to get banned posters: %s", err), http.StatusInternalServerError)
		return
	}

	err = bansT.Execute(w, banned)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to execute template: %s", err), http.StatusInternalServerError)
		return
	}
}
