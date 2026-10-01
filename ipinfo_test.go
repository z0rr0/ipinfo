// Copyright 2026 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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
	}{
		{name: "health", path: "/health", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "OK\n"},
		{name: "health_slash", path: "/health/", code: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "OK\n"},
		{name: "json_no_ip", path: "/json", code: http.StatusInternalServerError},
		{name: "json", path: "/json", ip: "193.138.218.226", code: http.StatusOK, contentType: "application/json; charset=utf-8"},
		{name: "root", path: "/", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8"},
		{name: "unknown", path: "/unknown", ip: "193.138.218.226", code: http.StatusOK, contentType: "text/plain; charset=utf-8"},
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

			if c.body == "" {
				return
			}

			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Fatal(readErr)
			}

			if strBody := string(body); strBody != c.body {
				t.Errorf("not equal body: %q", strBody)
			}
		})
	}
}
