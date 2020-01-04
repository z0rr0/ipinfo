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
)

// Cfg is configuration settings struct.
type Cfg struct {
	Host          string          `json:"host"`
	Port          uint            `json:"port"`
	Db            string          `json:"db"`
	IgnoreHeaders []string        `json:"ignore_headers"`
	IPHeader      string          `json:"ip_header"`
	CacheSize     int             `json:"cache_size"`
	ih            map[string]bool // ignored header map
	storage       *geoip2.Reader
	cache         *lru.Cache
}

// StrParam is common struct for headers and form params.
type StrParam struct {
	Name  string
	Value string
}

type stringParams []StrParam

func (a stringParams) Len() int           { return len(a) }
func (a stringParams) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stringParams) Less(i, j int) bool { return a[i].Name < a[j].Name }

// Addr returns service's net address.
func (c *Cfg) Addr() string {
	return net.JoinHostPort(c.Host, fmt.Sprint(c.Port))
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
	c.ih = make(map[string]bool)
	for _, h := range c.IgnoreHeaders {
		c.ih[strings.ToUpper(h)] = true
	}
	// db storage
	storage, err := geoip2.Open(c.Db)
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

// GetHeaders returns sorted request headers excluding ignored values.
func (c *Cfg) GetHeaders(r *http.Request) []StrParam {
	result := make([]StrParam, 0, len(r.Header))
	for k, v := range r.Header {
		if !c.ih[k] {
			// header is not ignored
			result = append(result, StrParam{k, strings.Join(v, "; ")})
		}
	}
	sort.Sort(stringParams(result))
	return result
}

// GetParams returns sorted request parameters.
func (c *Cfg) GetParams(r *http.Request) []StrParam {
	result := make([]StrParam, 0)
	r.FormValue("test") // init form load
	for k, v := range r.Form {
		result = append(result, StrParam{k, strings.Join(v, "; ")})
	}
	sort.Sort(stringParams(result))
	return result
}
