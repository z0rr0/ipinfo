// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/z0rr0/ipinfo/db"
)

// DbFile is storage configuration struct.
type DbFile struct {
	URL      string `json:"url"`
	File     string `json:"file"`
	CheckSum string `json:"checksum"`
	Format   string `json:"format"`
}

// Cfg is configuration settings struct.
type Cfg struct {
	Host          string   `json:"host"`
	Port          uint     `json:"port"`
	Db            DbFile   `json:"db"`
	IgnoreHeaders []string `json:"ignore_headers"`
	IPHeader      string   `json:"ip_header"`
	Storage       *geoip2.Reader
}

// Addr returns service's net address.
func (c *Cfg) Addr() string {
	return net.JoinHostPort(c.Host, fmt.Sprint(c.Port))
}

// IsIgnoredHeader return true is requested header is to be ignored.
func (c *Cfg) IsIgnoredHeader(h string) bool {
	l := len(c.IgnoreHeaders)
	if l == 0 {
		return false
	}
	header := strings.ToUpper(h)
	if i := sort.SearchStrings(c.IgnoreHeaders, header); i < l && c.IgnoreHeaders[i] == header {
		return true
	}
	return false
}

// GetIP return string IP address.
func (c *Cfg) GetIP(r *http.Request) (string, error) {
	if c.IPHeader == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", err
		}
		return host, nil
	}
	if values, ok := r.Header[c.IPHeader]; ok {
		return values[0], nil
	}
	return "", fmt.Errorf("not real ip header")
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
	sort.Strings(c.IgnoreHeaders)
	// uppercase
	for k, v := range c.IgnoreHeaders {
		c.IgnoreHeaders[k] = strings.ToUpper(v)
	}
	// db storage
	storage, err := db.GetDb(c.Db.URL, c.Db.File, c.Db.CheckSum, c.Db.Format)
	if err != nil {
		return nil, err
	}
	c.Storage = storage
	return c, nil
}
