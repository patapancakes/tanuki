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

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Port int `yaml:"port"`

	SiteName    string   `yaml:"siteName"`
	SiteSlogans []string `yaml:"siteSlogans"`

	AdminPassword string `yaml:"adminPassword"`
	AdminPostOnly bool   `yaml:"adminPostOnly"`

	PostCooldown  int `yaml:"postCooldown"` // in seconds
	AdminCooldown int `yaml:"adminCooldown"`

	MaxPostsPerPage int `yaml:"maxPostsPerPage"`
	MaxPages        int `yaml:"maxPages"`
	MaxBumps        int `yaml:"maxBumps"`

	MaxNameLength    int `yaml:"maxNameLength"`
	MaxSubjectLength int `yaml:"maxSubjectLength"`
	MaxCommentLength int `yaml:"maxCommentLength"`

	MaxUploadSize int `yaml:"maxUploadSize"` // in kilobytes

	ThumbnailDimensions int `yaml:"thumbnailDimensions"`
	ThumbnailQuality    int `yaml:"thumbnailQuality"`
}

var Config ConfigFile

func InitConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	err = yaml.NewDecoder(f).Decode(&Config)
	if err != nil {
		return err
	}

	return nil
}
