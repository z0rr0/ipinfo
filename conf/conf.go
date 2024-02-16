// Copyright 2023 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/oschwald/geoip2-golang"
)

const defaultISOCode = "en"

// Cfg is configuration settings struct.
type Cfg struct {
	Host           string              `json:"host"`
	Port           uint                `json:"port"`
	Db             string              `json:"db"`
	IgnoreHeaders  []string            `json:"ignore_headers"`
	IPHeader       string              `json:"ip_header"`
	CacheSize      int                 `json:"cache_size"`
	ignoredHeaders map[string]struct{} // ignored header map
	storage        *geoip2.Reader
	cache          *lru.Cache[string, *geoip2.City]
}

// StrParam is common struct for headers and form params.
type StrParam struct {
	Name  string
	Value string
}

// IPInfo is IP and related info for response.
type IPInfo struct {
	IP        string    `xml:"ip" json:"ip"`
	Country   string    `xml:"country" json:"country"`
	City      string    `xml:"city" json:"city"`
	Longitude float64   `xml:"longitude" json:"longitude"`
	Latitude  float64   `xml:"latitude" json:"latitude"`
	UTCTime   string    `xml:"utc_time" json:"utc_time"`
	TimeZone  string    `xml:"time_zone" json:"time_zone"`
	Language  string    `xml:"language" json:"language"`
	Timestamp time.Time `xml:"-" json:"-"`
}

// LocalTime returns local time in RFC3339 format or "-" if error.
func (i *IPInfo) LocalTime() string {
	loc, err := time.LoadLocation(i.TimeZone)
	if err != nil {
		return "-"
	}
	return i.Timestamp.In(loc).Format(time.RFC3339)
}

// LocalDateTime returns separated local date and time strings or "-" if error.
func (i *IPInfo) LocalDateTime() (string, string) {
	loc, err := time.LoadLocation(i.TimeZone)
	if err != nil {
		return "-", "-"
	}

	locTime := i.Timestamp.In(loc)
	return locTime.Format(time.DateOnly), locTime.Format(time.TimeOnly)
}

// Location returns location string.
func (i *IPInfo) Location() string {
	if i.Country == "" {
		return ""
	}
	if i.City == "" {
		return i.Country
	}
	return fmt.Sprintf("%s, %s", i.Country, i.City)
}

// Info returns base info about request.
func (c *Cfg) Info(r *http.Request) (*IPInfo, error) {
	host, err := c.GetIP(r)
	if err != nil {
		return nil, err
	}

	city, err := c.GetCity(host)
	if err != nil {
		return nil, err
	}

	isoCode := strings.ToLower(city.Country.IsoCode)
	if _, ok := city.Country.Names[isoCode]; !ok {
		isoCode = defaultISOCode
	}

	utcNow := time.Now().UTC()
	info := IPInfo{
		IP:        host,
		Country:   city.Country.Names[isoCode],
		City:      city.City.Names[isoCode],
		Longitude: city.Location.Longitude,
		Latitude:  city.Location.Latitude,
		UTCTime:   utcNow.Format(time.RFC3339),
		TimeZone:  city.Location.TimeZone,
		Language:  isoCode,
		Timestamp: utcNow,
	}
	return &info, nil
}

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
	return "", errors.New("no real ip header")
}

// GetCity returns city info found by IP address.
func (c *Cfg) GetCity(host string) (*geoip2.City, error) {
	if c.cache != nil {
		if city, ok := c.cache.Get(host); ok {
			return city, nil
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
	jsonData, err := readConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	c := &Cfg{}
	err = json.Unmarshal(jsonData, c)
	if err != nil {
		return nil, err
	}

	c.ignoredHeaders = make(map[string]struct{})
	for _, h := range c.IgnoreHeaders {
		c.ignoredHeaders[strings.ToUpper(h)] = struct{}{}
	}

	// db storage
	storage, err := geoip2.Open(c.Db)
	if err != nil {
		return nil, err
	}
	c.storage = storage

	if c.CacheSize > 0 {
		cache, err := lru.New[string, *geoip2.City](c.CacheSize)
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
		if _, ok := c.ignoredHeaders[strings.ToUpper(k)]; !ok {
			// header is not ignored
			result = append(result, StrParam{k, strings.Join(v, "; ")})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// GetParams returns sorted request parameters.
func (c *Cfg) GetParams(r *http.Request) []StrParam {
	result := make([]StrParam, 0)
	r.FormValue("test") // init form load
	for k, v := range r.Form {
		result = append(result, StrParam{k, strings.Join(v, "; ")})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func readConfig(filename string) ([]byte, error) {
	const (
		dockerDir  = "/data/conf"
		testConfig = "/tmp/ipinfo_test.json"
	)

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current dir: %w", err)
	}

	cleanPath := filepath.Clean(strings.Trim(filename, " "))

	if filepath.IsAbs(cleanPath) {
		if !(cleanPath == testConfig || strings.HasPrefix(cleanPath, dockerDir) || strings.HasPrefix(cleanPath, currentDir)) {
			return nil, fmt.Errorf("file %q has relative path and not in the allowed directories", cleanPath)
		}
	} else {

		cleanPath = filepath.Join(currentDir, cleanPath)
	}

	return os.ReadFile(cleanPath)
}
