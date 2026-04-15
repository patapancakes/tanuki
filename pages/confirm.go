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

	. "github.com/patapancakes/tanuki/db"
)

type ConfirmData struct {
	Post    Post
	Action  string
	Referer string
}

var confirmT *template.Template

func Confirm(w http.ResponseWriter, r *http.Request) {
	var cd ConfirmData

	err := checkAuth(r)
	if err != nil {
		writeError(w, r, fmt.Sprintf("authentication failed: %s", err), http.StatusUnauthorized)
		return
	}

	cd.Post, err = posts.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to fetch post: %s", err), http.StatusInternalServerError)
		return
	}

	cd.Action = r.PathValue("action")
	cd.Referer = r.Referer()

	err = confirmT.Execute(w, cd)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to execute template: %s", err), http.StatusInternalServerError)
		return
	}
}
