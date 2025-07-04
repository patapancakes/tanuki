/*
	tanuki - a lightweight image bbs
	Copyright (C) 2024  Pancakes (pancakes@mooglepowered.com)

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
	"html/template"
	"math/rand/v2"

	. "github.com/patapancakes/tanuki/config"

	"github.com/xeonx/timeago"
)

var funcs = template.FuncMap{
	"timeago": timeago.English.Format,
	"sum":     func(a, b int) int { return a + b },
	"sub":     func(a, b int) int { return a - b },
	"max":     func(a, b int) int { return max(a, b) },
	"config":  func() ConfigFile { return Config },
	"rand":    rand.IntN,
}

func Init() error {
	var err error

	// home
	homeT, err = template.New("home.html").Funcs(funcs).ParseFiles("data/templates/home.html")
	if err != nil {
		return err
	}

	homeT, err = homeT.ParseGlob("data/templates/include/*.html")
	if err != nil {
		return err
	}

	// thread
	threadT, err = template.New("thread.html").Funcs(funcs).ParseFiles("data/templates/thread.html")
	if err != nil {
		return err
	}

	threadT, err = threadT.ParseGlob("data/templates/include/*.html")
	if err != nil {
		return err
	}

	// error
	errorT, err = template.New("error.html").Funcs(funcs).ParseFiles("data/templates/error.html")
	if err != nil {
		return err
	}

	errorT, err = errorT.ParseGlob("data/templates/include/*.html")
	if err != nil {
		return err
	}

	// admin
	adminT, err = template.New("admin.html").Funcs(funcs).ParseFiles("data/templates/admin.html")
	if err != nil {
		return err
	}

	adminT, err = adminT.ParseGlob("data/templates/include/*.html")
	if err != nil {
		return err
	}

	return nil
}
