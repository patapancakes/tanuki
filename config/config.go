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

package config

import (
	"encoding/json"
	"os"
)

type ConfigFile struct {
	Port int `json:"port"`

	SiteName    string   `json:"siteName"`
	SiteSlogans []string `json:"siteSlogans"`

	AdminPassword string `json:"adminPassword"`

	PostCooldown  int `json:"postCooldown"` // in seconds
	AdminCooldown int `json:"adminCooldown"`

	MaxPostsPerPage int `json:"maxPostsPerPage"`
	MaxPages        int `json:"maxPages"`
	MaxBumps        int `json:"maxBumps"`

	MaxNameLength    int `json:"maxNameLength"`
	MaxSubjectLength int `json:"maxSubjectLength"`
	MaxCommentLength int `json:"maxCommentLength"`

	MaxUploadSize int `json:"maxUploadSize"` // in kilobytes

	ThumbnailDimensions int `json:"thumbnailDimensions"`
	ThumbnailQuality    int `json:"thumbnailQuality"`
}

var Config ConfigFile

func InitConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		return err
	}

	return nil
}
