// Copyright 2026 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/z0rr0/ipinfo/conf"
	"github.com/z0rr0/ipinfo/handle"
)

const testConfigName = "/tmp/ipinfo_test.json"

func TestNewHandler(t *testing.T) {
	cfg, err := conf.New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			t.Errorf("close error: %v", closeErr)
		}
	}()

	handler := newHandler(cfg, &handle.BuildInfo{})

	cases := []struct {
		name        string
		path        string
		ip          string
		code        int
		contentType string
		body        string
		contains    string
	}{
		{name: "health", path: "/health", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "OK\n"},
		{name: "health_slash", path: "/health/", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "OK\n"},
		{name: "json_no_ip", path: "/json", code: http.StatusInternalServerError},
		{name: "json", path: "/json", ip: "193.138.218.226", code: http.StatusOK, contentType: "application/json; charset=utf-8"},
		{name: "ip", path: "/ip", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "193.138.218.226\n"},
		{name: "ip_slash", path: "/ip/", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "193.138.218.226\n"},
		{name: "root", path: "/", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8"},
		{name: "unknown", path: "/unknown", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8"},
		{
			name: "json_param", path: "/json?ip=193.138.218.226", code: http.StatusOK,
			contentType: "application/json; charset=utf-8", contains: `"ip":"193.138.218.226"`,
		},
		{name: "ip_param", path: "/ip?ip=8.8.8.8", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "8.8.8.8\n"},
		{name: "json_bad_param", path: "/json?ip=bad", code: http.StatusBadRequest, body: "invalid ip parameter\n"},
		{name: "health_bad_param", path: "/health?ip=bad", code: http.StatusOK, body: "OK\n"},
		{name: "version_no_ip", path: "/version", code: http.StatusOK, contentType: "text/plain; charset=utf-8", contains: "Version:"},
		{name: "version_bad_param", path: "/version?ip=bad", code: http.StatusOK, contains: "Version:"},
		{
			name: "full_param", path: "/full?ip=8.8.8.8", code: http.StatusOK,
			contentType: "text/html; charset=utf-8", contains: `href="/json?ip=8.8.8.8"`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "https://example.com"+c.path, nil)
			if c.ip != "" {
				req.Header.Add("X-Real-Ip", c.ip)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					t.Error(closeErr)
				}
			}()

			if resp.StatusCode != c.code {
				t.Errorf("not %d status code: %v", c.code, resp.StatusCode)
			}

			if ct := resp.Header.Get("Content-Type"); c.contentType != "" && ct != c.contentType {
				t.Errorf("not equal Content-Type: %v", ct)
			}

			if c.body == "" && c.contains == "" {
				return
			}

			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Fatal(readErr)
			}

			strBody := string(body)
			if c.body != "" && strBody != c.body {
				t.Errorf("not equal body: %q", strBody)
			}

			if !strings.Contains(strBody, c.contains) {
				t.Errorf("body %q does not contain %q", strBody, c.contains)
			}
		})
	}
}
