package handle

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/z0rr0/ipinfo/conf"
)

const testConfigName = "/tmp/ipinfo_test.json"

func TestJSONHandler(t *testing.T) {
	cfg, err := conf.New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Add("X-Real-Ip", "193.138.218.226")

	w := httptest.NewRecorder()
	err = JSONHandler(w, req, cfg)
	if err != nil {
		t.Fatal(err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("not %d status code: %v", http.StatusOK, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("not equal Content-Type: %v", ct)
	}

	info := &JSONInfo{}
	err = json.NewDecoder(resp.Body).Decode(info)
	if err != nil {
		t.Fatal(err)
	}

	expected := &JSONInfo{
		IP:        "193.138.218.226",
		Country:   "Sweden",
		City:      "Malmo",
		Longitude: 12.9982,
		Latitude:  55.6078,
		UTCTime:   info.UTCTime, // don't check time
	}
	if *info != *expected {
		t.Errorf("not equal JSONInfo: %v", info)
	}

	if err = resp.Body.Close(); err != nil {
		t.Error(err)
	}
}

func TestTextShortHandler(t *testing.T) {
	cfg, err := conf.New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Add("X-Real-Ip", "193.138.218.226")

	w := httptest.NewRecorder()
	err = TextShortHandler(w, req, cfg)
	if err != nil {
		t.Fatal(err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("not %d status code: %v", http.StatusOK, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("not equal Content-Type: %v", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	strBody := string(body)
	i := strings.Index(strBody, "TimeUTC")
	if i < 0 {
		t.Fatalf("not found TimeUTC: %v", strBody)
	}

	strBody = strBody[:i]
	expected := "IP:      193.138.218.226\nCountry: Sweden\nCity:    Malmo\n"
	if strBody != expected {
		t.Errorf("not equal text body: %v", strBody)
	}
}

func TestTextHandler(t *testing.T) {
	cfg, err := conf.New(testConfigName)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "http://example.com/foo?b=1&c=3", nil)
	req.Header.Add("X-Real-Ip", "193.138.218.226")
	req.Header.Add("X-Header-A", "a")

	w := httptest.NewRecorder()
	err = TextHandler(w, req, cfg)
	if err != nil {
		t.Fatal(err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("not %d status code: %v", http.StatusOK, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("not equal Content-Type: %v", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	strBody := string(body)
	subStr := "IP: 193.138.218.226\nProto: HTTP/1.1\nMethod: GET"

	if !strings.Contains(strBody, subStr) {
		t.Fatalf("not found required first sub-string: %v", strBody)
	}

	subStr = "Locations\n---------\nCountry: Sweden\nCity: Malmo\nLatitude: 55.6078\nLongitude: 12.9982\nTimeZone:"
	if !strings.Contains(strBody, subStr) {
		t.Fatalf("not found required second sub-string: %v", strBody)
	}
}
