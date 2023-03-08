package handle

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/z0rr0/ipinfo/conf"
)

// printF is fmt.Fprintf wrapper with error check.
func printF(err error, w io.Writer, format string, a ...interface{}) error {
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, format, a...)
	return err
}

func sectionHeadersAndParams(err error, w io.Writer, r *http.Request, cfg *conf.Cfg) error {
	if err != nil {
		return err
	}

	for _, h := range cfg.GetHeaders(r) {
		err = printF(err, w, "%v: %v\n", h.Name, h.Value)
	}

	err = printF(err, w, "\nParams\n---------\n")
	for _, p := range cfg.GetParams(r) {
		err = printF(err, w, "%v: %v\n", p.Name, p.Value)
	}
	return err
}

func sectionLocation(err error, host string, w io.Writer, cfg *conf.Cfg) error {
	if err != nil {
		return err
	}

	city, err := cfg.GetCity(host)
	err = printF(err, w, "\nLocations\n---------\n")
	if err != nil {
		return err
	}

	isoCode := strings.ToLower(city.Country.IsoCode)
	if _, ok := city.Country.Names[isoCode]; !ok {
		isoCode = defaultISOCode
	}

	err = printF(err, w, "Country: %v\n", city.Country.Names[isoCode])
	err = printF(err, w, "City: %v\n", city.City.Names[isoCode])
	err = printF(err, w, "Latitude: %v\n", city.Location.Latitude)
	err = printF(err, w, "Longitude: %v\n", city.Location.Longitude)
	err = printF(err, w, "TimeZone: %v\n", city.Location.TimeZone)
	return printF(err, w, "TimeUTC: %v\n", time.Now().UTC().Format(time.RFC3339))
}

func sectionShortLocation(err error, host string, w io.Writer, cfg *conf.Cfg) error {
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

	err = printF(err, w, "Country: %v\n", city.Country.Names[isoCode])
	err = printF(err, w, "City:    %v\n", city.City.Names[isoCode])
	return printF(err, w, "TimeUTC: %v\n", time.Now().UTC().Format(time.RFC3339))
}
