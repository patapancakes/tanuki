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
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	. "github.com/patapancakes/tanuki/config"
	"github.com/patapancakes/tanuki/pages"
)

func main() {
	log.Printf("Tanuki BBS by Pancakes (pancakes@mooglepowered.com)\n")
	log.Printf("https://github.com/patapancakes/tanuki\n")

	configpath := flag.String("config", "config.yml", "path to config file")
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

	// session keys
	err = checkKey()
	if err != nil {
		log.Fatalf("failed to create session keys: %s", err)
	}

	// files
	http.Handle("GET /assets/", cache(http.StripPrefix("/assets/", http.FileServerFS(pages.AssetsFS))))
	http.Handle("GET /thumb/", cache(http.StripPrefix("/thumb/", http.FileServer(http.Dir("data/thumb")))))
	http.Handle("GET /full/", cache(http.StripPrefix("/full/", http.FileServer(http.Dir("data/full")))))

	http.HandleFunc("GET /", pages.Home)
	http.HandleFunc("GET /{page}", pages.Home)

	http.HandleFunc("GET /thread/{id}", pages.Thread)

	http.HandleFunc("GET /admin", pages.Admin)
	http.HandleFunc("GET /admin/bans", pages.Bans)

	http.HandleFunc("POST /admin/login", pages.AdminLogin)
	http.HandleFunc("GET /admin/logout", pages.AdminLogout)

	http.HandleFunc("GET /admin/delete/{id}", pages.AdminDelete)
	http.HandleFunc("GET /admin/ban/{id}", pages.AdminBan)

	http.HandleFunc("POST /admin/unbanid", pages.AdminUnbanID)

	http.HandleFunc("POST /newpost", pages.NewPost)

	log.Printf("now listening on port %d", Config.Port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func checkKey() error {
	f, err := os.OpenFile("data/session.key", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}

		return err
	}

	_, err = io.Copy(f, io.LimitReader(rand.Reader, 32))
	if err != nil {
		return err
	}

	return nil
}

func cache(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Expires", time.Now().Add(time.Hour*24*7).Format(time.RFC1123))
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		h.ServeHTTP(w, r)
	}
}
