// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/z0rr0/ipinfo/db"
)

// DbFile is storage configuration struct.
type DbFile struct {
	URL      string `json:"url"`
	File     string `json:"file"`
	CheckSum string `json:"checksum"`
}

// Cfg is configuration settings struct.
type Cfg struct {
	Host    string `json:"host"`
	Port    uint   `json:"port"`
	Db      DbFile `json:"db"`
	Storage *geoip2.Reader
}

// New returns new rates configuration.
func New(filename string) (*Cfg, error) {
	fullPath, err := filepath.Abs(strings.Trim(filename, " "))
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	jsonData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	c := &Cfg{}
	err = json.Unmarshal(jsonData, c)
	if err != nil {
		return nil, err
	}
	storage, err := db.GetDb(c.Db.URL, c.Db.File, c.Db.CheckSum)
	if err != nil {
		return nil, err
	}
	c.Storage = storage
	return c, nil
}
