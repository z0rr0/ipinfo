package handle

import (
	"net/http"

	"github.com/z0rr0/ipinfo/conf"
)

const defaultISOCode = "en"

// TextHandler is handler for text/plain response.
func TextHandler(w http.ResponseWriter, r *http.Request, cfg *conf.Cfg) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	host, err := cfg.GetIP(r)

	err = printF(err, w, "IP: %v\nProto: %v\nMethod: %v\nURI: %v\n", host, r.Proto, r.Method, r.RequestURI)
	err = printF(err, w, "\nHeaders\n---------\n")

	err = sectionHeadersAndParams(err, w, r, cfg)
	return sectionLocation(err, host, w, cfg)
}

// TextShortHandler is handler for text/plain response with short info.
// It returns only IP address, country, city and time.
func TextShortHandler(w http.ResponseWriter, r *http.Request, cfg *conf.Cfg) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	host, err := cfg.GetIP(r)

	err = printF(err, w, "IP:      %v\n", host)
	return sectionShortLocation(err, host, w, cfg)
}

func JSONHandler(w http.ResponseWriter, r *http.Request, cfg *conf.Cfg) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return sectionJSON(w, r, cfg)
}
