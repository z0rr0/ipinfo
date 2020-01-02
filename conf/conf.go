// Copyright 2020 Alexander Zaytsev <thebestzorro@yandex.ru>.
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

	"github.com/hashicorp/golang-lru"
	"github.com/oschwald/geoip2-golang"
	"github.com/z0rr0/ipinfo/db"
)

// Cfg is configuration settings struct.
type Cfg struct {
	Host          string   `json:"host"`
	Port          uint     `json:"port"`
	Db            string   `json:"db"`
	IgnoreHeaders []string `json:"ignore_headers"`
	IPHeader      string   `json:"ip_header"`
	CacheSize     int      `json:"cache_size"`
	storage       *geoip2.Reader
	cache         *lru.Cache
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

// GetCity returns city info found by IP address.
func (c *Cfg) GetCity(host string) (*geoip2.City, error) {
	if c.cache != nil {
		if v, ok := c.cache.Get(host); ok {
			return v.(*geoip2.City), nil
		}
	}
	ip := net.ParseIP(host)
	city, err := c.storage.City(ip)
	if err != nil {
		return nil, err
	}
	if c.cache != nil {
		c.cache.Add(host, city)
	}
	return city, nil
}

// Close closes db storage file.
func (c *Cfg) Close() error {
	if c.storage != nil {
		return c.storage.Close()
	}
	return nil
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
	storage, err := db.GetDb(c.Db)
	if err != nil {
		return nil, err
	}
	c.storage = storage
	if c.CacheSize > 0 {
		cache, err := lru.New(c.CacheSize)
		if err != nil {
			return nil, err
		}
		c.cache = cache
	}
	return c, nil
}
