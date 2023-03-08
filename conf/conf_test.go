// Copyright 2023 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"net/http/httptest"
	"testing"
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

	cases := []struct {
		name       string
		ipHeader   string
		ipValues   []string
		remoteAddr string
		ipAddress  string // error if empty
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
