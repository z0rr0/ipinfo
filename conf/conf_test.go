// Copyright 2020 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package conf contains methods and structures for configuration.
package conf

import (
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testConfigName = "/tmp/ipinfo_test.json"
)

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
	err = cfg.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
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
}

func TestCfg_IsIgnoredHeader(t *testing.T) {
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(cfg.IgnoreHeaders); l == 0 {
		t.Error("empty ignore headers list")
	}
	if _, ok := cfg.ignoredHeaders["BAD_HEADER"]; ok {
		t.Error("expected false for bad header value")
	}
	if _, ok := cfg.ignoredHeaders[strings.ToUpper(cfg.IgnoreHeaders[0])]; !ok {
		t.Error("expected true for valid header value")
	}
}

func TestCfg_GetIP(t *testing.T) {
	ipStr := "127.0.0.123"
	cfg, err := New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "https://example.com/foo", nil)
	_, err = cfg.GetIP(req)
	if err == nil {
		t.Error("expected 'not real ip header' error")
	}
	req.Header[cfg.IPHeader] = []string{ipStr}
	ip, err := cfg.GetIP(req)
	if err != nil {
		t.Error(err)
	}
	if ip != ipStr {
		t.Errorf("not equal %v != %v", ip, ipStr)
	}
}
