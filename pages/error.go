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
	"log"
	"net/http"
	"net/netip"
)

var errorT *template.Template

func writeError(w http.ResponseWriter, r *http.Request, error string, code int) {
	w.WriteHeader(code)

	errorT.Execute(w, error)

	addr := "unknown"
	addrport, err := netip.ParseAddrPort(r.RemoteAddr)
	if err == nil {
		addr = addrport.Addr().String()
	}

	log.Printf("[%s] %s: %s", addr, r.URL.Path, error)
}
