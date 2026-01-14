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
	"math"
	"net/http"
	"strconv"

	. "github.com/patapancakes/tanuki/config"
	. "github.com/patapancakes/tanuki/db"
)

type HomeData struct {
	Admin bool

	Posts PostData

	Page  int
	Pages int
}

var homeT *template.Template

func Home(w http.ResponseWriter, r *http.Request) {
	var hd HomeData
	var err error

	if Config.AdminPassword != "" {
		adminpw, err := r.Cookie("adminpw")
		if err != nil {
			if err != http.ErrNoCookie {
				writeError(w, r, fmt.Sprintf("failed to read admin password cookie: %s", err), http.StatusBadRequest)
				return
			}
		} else {
			hd.Admin = adminpw.Value == Config.AdminPassword
		}
	}

	hd.Page = 1
	if r.PathValue("page") != "" {
		hd.Page, _ = strconv.Atoi(r.PathValue("page"))
	}
	if hd.Page < 1 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	hd.Posts, err = posts.GetAll()
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to fetch posts: %s", err), http.StatusInternalServerError)
		return
	}

	hd.Pages = 1
	if len(hd.Posts) > Config.MaxPostsPerPage {
		hd.Pages = int(math.Ceil(float64(len(hd.Posts)) / float64(Config.MaxPostsPerPage)))
	}
	if hd.Pages < hd.Page {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	hd.Posts = hd.Posts[min((hd.Page-1)*Config.MaxPostsPerPage, len(hd.Posts)):]
	hd.Posts = hd.Posts[:min(Config.MaxPostsPerPage, len(hd.Posts))]

	err = homeT.Execute(w, hd)
	if err != nil {
		writeError(w, r, fmt.Sprintf("failed to execute template: %s", err), http.StatusInternalServerError)
		return
	}
}
