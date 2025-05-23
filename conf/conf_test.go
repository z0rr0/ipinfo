// Copyright 2023 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"net/http/httptest"
	"testing"
	"time"
)

const testConfigName = "/tmp/ipinfo_test.json"

func TestNew(t *testing.T) {
	if _, err := New("/bad_file_path.json"); err == nil {
		t.Error("unexpected behavior")
	}

	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Addr() == "" {
		t.Error("empty address")
	}

	if err = cfg.Close(); err != nil {
		t.Errorf("close error: %v", err)
	}

	cfg.storage = nil
	if err = cfg.Close(); err != nil {
		t.Errorf("close error with empty storage: %v", err)
	}
}

func TestCfg_GetCity(t *testing.T) {
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	_, err = cfg.GetCity("127.0.0.1")
	if err != nil {
		t.Errorf("get city error: %v", err)
	}

	// read from cache
	_, err = cfg.GetCity("127.0.0.1")
	if err != nil {
		t.Errorf("get city error: %v", err)
	}

	if _, ok := cfg.cache.Get("127.0.0.1"); !ok {
		t.Error("cache miss")
	}
}

func TestCfg_GetIP(t *testing.T) {
	cfg, e := New(testConfigName)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	cases := []struct {
		name       string
		ipHeader   string
		remoteAddr string
		ipAddress  string
		ipValues   []string
	}{
		{
			name:     "has header, no values",
			ipHeader: cfg.IPHeader,
		},
		{
			name:       "no header, bad remote addr",
			remoteAddr: "bad ip",
		},
		{
			name:       "no header, good remote addr",
			remoteAddr: "127.0.0.123:8082",
			ipAddress:  "127.0.0.123",
		},
		{
			name:      "has header and values",
			ipHeader:  cfg.IPHeader,
			ipValues:  []string{"127.0.0.123"},
			ipAddress: "127.0.0.123",
		},
	}

	req := httptest.NewRequest("GET", "https://example.com/foo", nil)
	for _, c := range cases {
		cfg.IPHeader = c.ipHeader
		if len(c.ipValues) > 0 {
			req.Header[c.ipHeader] = c.ipValues
		}
		req.RemoteAddr = c.remoteAddr

		ip, err := cfg.GetIP(req)
		if c.ipAddress == "" {
			// error expected
			if err == nil {
				t.Errorf("%s: expected error", c.name)
			}
		} else {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", c.name, err)
			}
			if ip != c.ipAddress {
				t.Errorf("%s: not equal %v != %v", c.name, ip, c.ipAddress)
			}
		}
	}
}

func TestCfg_GetHeaders(t *testing.T) {
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	req := httptest.NewRequest("GET", "https://example.com/foo", nil)
	result := cfg.GetHeaders(req)
	if len(result) > 0 {
		t.Errorf("not empty result: %v", result)
	}

	req.Header.Add("X-Header-B", "b")
	req.Header.Add("X-Header-C", "c")
	req.Header.Add("X-Header-A", "a")

	result = cfg.GetHeaders(req)
	if result == nil {
		t.Fatal("empty result")
	}

	expected := []StrParam{
		{Name: "X-Header-A", Value: "a"},
		{Name: "X-Header-B", Value: "b"},
		{Name: "X-Header-C", Value: "c"},
	}
	if len(result) != len(expected) {
		t.Errorf("not equal length: %v != %v", len(result), len(expected))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("not equal %v != %v", result[i], expected[i])
		}
	}
}

func TestCfg_GetParams(t *testing.T) {
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	req := httptest.NewRequest("GET", "https://example.com/foo", nil)
	result := cfg.GetParams(req)
	if len(result) > 0 {
		t.Errorf("not empty result: %v", result)
	}

	req = httptest.NewRequest("GET", "https://example.com/foo?c=3&b=2&a=1", nil)
	result = cfg.GetParams(req)
	if result == nil {
		t.Fatal("empty result")
	}

	expected := []StrParam{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
		{Name: "c", Value: "3"},
	}
	if len(result) != len(expected) {
		t.Errorf("not equal length: %v != %v", len(result), len(expected))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("not equal %v != %v", result[i], expected[i])
		}
	}
}

func TestCfg_Info(t *testing.T) {
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	req := httptest.NewRequest("GET", "https://example.com/foo", nil)
	req.Header.Add("X-Real-Ip", "193.138.218.226")

	info, err := cfg.Info(req)
	if err != nil {
		t.Fatalf("info error: %v", err)
	}

	expected := IPInfo{
		IP:        "193.138.218.226",
		Country:   "Sweden",
		City:      "Malmo",
		Longitude: 12.9982,
		Latitude:  55.6078,
		TimeZone:  "Europe/Stockholm",
		Language:  defaultISOCode,
		// don't check time fields
		UTCTime:   info.UTCTime,
		Timestamp: info.Timestamp,
	}

	if i := *info; i != expected {
		t.Errorf("not equal %v != %v", i, expected)
	}
}

func TestIPInfo_LocalTime(t *testing.T) {
	var info IPInfo
	ts := time.Date(2019, 1, 2, 3, 4, 5, 6, time.UTC)
	cases := []struct {
		name     string
		timeZone string
		expected string
	}{
		{name: "empty", expected: "2019-01-02T03:04:05Z"},
		{name: "invalid", timeZone: "invalid", expected: "-"},
		{name: "Europe/Stockholm", timeZone: "Europe/Stockholm", expected: "2019-01-02T04:04:05+01:00"},
	}
	for _, c := range cases {
		info = IPInfo{TimeZone: c.timeZone, Timestamp: ts}
		if result := info.LocalTime(); result != c.expected {
			t.Errorf("%s: not equal %v != %v", c.name, result, c.expected)
		}
	}
}

func TestIPInfo_Location(t *testing.T) {
	var info IPInfo
	cases := []struct {
		name     string
		country  string
		city     string
		expected string
	}{
		{name: "empty", expected: ""},
		{name: "country", country: "Sweden", expected: "Sweden"},
		{name: "city", city: "Malmo", expected: ""},
		{name: "country and city", country: "Sweden", city: "Malmo", expected: "Sweden, Malmo"},
	}
	for _, c := range cases {
		info = IPInfo{Country: c.country, City: c.city}
		if result := info.Location(); result != c.expected {
			t.Errorf("%s: not equal %v != %v", c.name, result, c.expected)
		}
	}
}
