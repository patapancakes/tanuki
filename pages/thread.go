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

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"
)

var threadT *template.Template

type ThreadData struct {
	Admin bool

	Post Post
}

func Thread(w http.ResponseWriter, r *http.Request) {
	var td ThreadData

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
			td.Admin = true
		}
	}

	var err error
	td.Post, err = posts.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to fetch post: %s", err), http.StatusInternalServerError)
		return
	}
	if !td.Post.IsThread() {
		http.Redirect(w, r, fmt.Sprintf("/thread/%s#post_%s", td.Post.Parent, td.Post.ID()), http.StatusSeeOther)
		return
	}

	err = threadT.Execute(w, td)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to execute template: %s", err), http.StatusInternalServerError)
		return
	}
}
