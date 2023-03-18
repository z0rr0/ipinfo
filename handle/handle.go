package handle

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"

	"github.com/z0rr0/ipinfo/conf"
)

var (
	//go:embed index.html
	htmlIndex    string
	htmlTemplate = template.Must(template.New("index").Parse(htmlIndex))
)

// BuildInfo is a struct for version information.
type BuildInfo struct {
	Version   string
	Revision  string
	BuildDate string
	GoVersion string
}

// String returns a string representation of BuildInfo.
func (b *BuildInfo) String() string {
	return fmt.Sprintf(
		"\tVersion: %v\n\tRevision: %v\n\tBuild date: %v\n\tGo version: %v",
		b.Version, b.Revision, b.BuildDate, b.GoVersion,
	)
}

// XMLInfo is a struct for application/xml response data.
type XMLInfo struct {
	XMLName xml.Name `xml:"ipinfo"`
	conf.IPInfo
}

// TextHandler is handler for text/plain response.
func TextHandler(w http.ResponseWriter, r *http.Request, cfg *conf.Cfg, info *conf.IPInfo) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	err := printF(nil, w, "IP: %v\nProto: %v\nMethod: %v\nURI: %v\n",
		info.IP, r.Proto, r.Method, r.RequestURI,
	)
	err = printF(err, w, "\nHeaders\n---------\n")

	err = sectionHeadersAndParams(err, w, r, cfg)
	return sectionLocation(err, w, info)
}

// TextShortHandler is handler for text/plain response with short info.
// It returns only IP address, country, city and time.
func TextShortHandler(w http.ResponseWriter, info *conf.IPInfo, _ *BuildInfo) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	err := printF(nil, w, "IP:      %v\n", info.IP)
	err = printF(err, w, "Country: %v\n", info.Country)
	err = printF(err, w, "City:    %v\n", info.City)
	return printF(err, w, "TimeUTC: %v\n", info.UTCTime)
}

// JSONHandler is handler for application/json response.
func JSONHandler(w http.ResponseWriter, info *conf.IPInfo, _ *BuildInfo) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(info)
}

// XMLHandler is handler for application/xml response.
func XMLHandler(w http.ResponseWriter, info *conf.IPInfo, _ *BuildInfo) error {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if err := printF(nil, w, xml.Header); err != nil {
		return fmt.Errorf("XMLHandler: %w", err)
	}
	return xml.NewEncoder(w).Encode(&XMLInfo{IPInfo: *info})
}

// HTMLHandler is handler for text/html response.
func HTMLHandler(w http.ResponseWriter, info *conf.IPInfo, _ *BuildInfo) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return htmlTemplate.Execute(w, info)
}

// VersionHandler is handler for version information.
func VersionHandler(w http.ResponseWriter, info *conf.IPInfo, buildInfo *BuildInfo) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	err := printF(nil, w, "Version:    %v\n", buildInfo.Version)
	err = printF(err, w, "Revision:   %v\n", buildInfo.Revision)
	err = printF(err, w, "Build date: %v\n", buildInfo.BuildDate)
	err = printF(err, w, "Go version: %v\n", buildInfo.GoVersion)
	err = printF(err, w, "Language:   %v\n", info.Language)
	err = printF(err, w, "Location:   %v\n", info.Location())
	return printF(err, w, "Local time: %v\n", info.LocalTime())
}
