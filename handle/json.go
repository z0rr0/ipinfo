package handle

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/z0rr0/ipinfo/conf"
)

// JSONInfo is a struct for application/json response data.
type JSONInfo struct {
	IP        string  `json:"ip"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	UTCTime   string  `json:"utc_time"`
}

func sectionJSON(w io.Writer, r *http.Request, cfg *conf.Cfg) error {
	host, err := cfg.GetIP(r)
	if err != nil {
		return err
	}

	city, err := cfg.GetCity(host)
	if err != nil {
		return err
	}

	isoCode := strings.ToLower(city.Country.IsoCode)
	if _, ok := city.Country.Names[isoCode]; !ok {
		isoCode = defaultISOCode
	}

	info := &JSONInfo{
		IP:        host,
		Country:   city.Country.Names[isoCode],
		City:      city.City.Names[isoCode],
		Longitude: city.Location.Longitude,
		Latitude:  city.Location.Latitude,
		UTCTime:   time.Now().UTC().Format(time.RFC3339),
	}
	return json.NewEncoder(w).Encode(info)
}
