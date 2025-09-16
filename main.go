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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	. "github.com/patapancakes/tanuki/config"
	"github.com/patapancakes/tanuki/pages"
)

func main() {
	log.Printf("Tanuki BBS by Pancakes (pancakes@mooglepowered.com)\n")
	log.Printf("https://github.com/patapancakes/tanuki\n")

	configpath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	err := InitConfig(*configpath)
	if err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}

	// templates
	err = pages.Init()
	if err != nil {
		log.Fatalf("failed to create page templates: %s", err)
	}

	// create directories
	os.MkdirAll("data/thumb", 0755)
	os.MkdirAll("data/full", 0755)

	// files
	http.Handle("GET /data/static/", http.StripPrefix("/data/static/", http.FileServer(http.Dir("data/static"))))
	http.Handle("GET /data/thumb/", http.StripPrefix("/data/thumb/", http.FileServer(http.Dir("data/thumb"))))
	http.Handle("GET /data/full/", http.StripPrefix("/data/full/", http.FileServer(http.Dir("data/full"))))

	// routes
	http.HandleFunc("GET /", pages.Home)
	http.HandleFunc("GET /{page}", pages.Home)

	http.HandleFunc("GET /thread/{id}", pages.Thread)

	http.HandleFunc("GET /admin", pages.Admin)

	http.HandleFunc("POST /admin/login", pages.AdminLogin)
	http.HandleFunc("GET /admin/logout", pages.AdminLogout)

	http.HandleFunc("GET /admin/delete/{id}", pages.AdminDelete)
	http.HandleFunc("GET /admin/ban/{id}", pages.AdminBan)

	http.HandleFunc("POST /newpost", pages.NewPost)

	log.Printf("now listening on port %d", Config.Port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
