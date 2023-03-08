package handle

import (
	"fmt"
	"io"
	"net/http"

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

func sectionLocation(err error, w io.Writer, info *conf.IPInfo) error {
	if err != nil {
		return err
	}

	err = printF(err, w, "\nLocations\n---------\n")
	if err != nil {
		return err
	}

	err = printF(err, w, "Country: %v\n", info.Country)
	err = printF(err, w, "City: %v\n", info.City)
	err = printF(err, w, "Latitude: %v\n", info.Latitude)
	err = printF(err, w, "Longitude: %v\n", info.Longitude)
	err = printF(err, w, "TimeZone: %v\n", info.TimeZone)
	return printF(err, w, "TimeUTC: %v\n", info.UTCTime)
}
