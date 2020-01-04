// Copyright 2020 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package main implements main method of IPINFO service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/z0rr0/ipinfo/conf"
)

const (
	// Name is a program name
	Name = "IPINFO"
	// Config is default configuration file name
	Config  = "config.json"
	timeout = 30 * time.Second
)

var (
	// Version is program git version
	Version = ""
	// Revision is revision number
	Revision = ""
	// BuildDate is build date
	BuildDate = ""
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()

	// internal logger
	loggerInfo = log.New(os.Stdout, fmt.Sprintf("INFO [%v]: ", Name),
		log.Ldate|log.Ltime|log.Lshortfile)
)

// IsError checks err error, writes its response and returns true if a problem was.
func IsError(w http.ResponseWriter, err error) (int, bool) {
	if err != nil {
		loggerInfo.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return http.StatusInternalServerError, true
	}
	return http.StatusOK, false
}

// WriteResult is fmt.Fprintf wrapper with error check.
func WriteResult(err error, w http.ResponseWriter, format string, a ...interface{}) error {
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, format, a...)
	return err
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			loggerInfo.Printf("abnormal termination [%v]: \n\t%v\n", Version, r)
		}
	}()
	version := flag.Bool("version", false, "show version")
	config := flag.String("config", Config, "configuration file")
	flag.Parse()

	versionInfo := fmt.Sprintf("\tVersion: %v\n\tRevision: %v\n\tBuild date: %v\n\tGo version: %v",
		Version, Revision, BuildDate, GoVersion)
	if *version {
		fmt.Println(versionInfo)
		return
	}

	cfg, err := conf.New(*config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			loggerInfo.Printf("cfg close error: %v\n", err)
		}
	}()
	srv := &http.Server{
		Addr:           cfg.Addr(),
		Handler:        http.DefaultServeMux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20, // 1MB
		ErrorLog:       loggerInfo,
	}
	loggerInfo.Printf("\n%v\nlisten addr: %v\n", versionInfo, srv.Addr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start, code, failed := time.Now(), http.StatusOK, false
		defer func() {
			loggerInfo.Printf("%-5v %v\t%-12v\t%v",
				r.Method,
				code,
				time.Since(start),
				r.RemoteAddr,
			)
		}()
		host, err := cfg.GetIP(r)
		// main info
		err = WriteResult(err, w, "IP: %v\nProto: %v\nMethod: %v\nURI: %v\n", host, r.Proto, r.Method, r.RequestURI)
		err = WriteResult(err, w, "\nHeaders\n---------\n")
		code, failed = IsError(w, err)
		if failed {
			return
		}
		// headers values
		for _, h := range cfg.GetHeaders(r) {
			err = WriteResult(err, w, "%v: %v\n", h.Name, h.Value)
		}
		err = WriteResult(err, w, "\nParams\n---------\n")
		code, failed = IsError(w, err)
		if failed {
			return
		}
		for _, p := range cfg.GetParams(r) {
			err = WriteResult(err, w, "%v: %v\n", p.Name, p.Value)
		}
		code, failed = IsError(w, err)
		if failed {
			return
		}
		// additional info
		city, err := cfg.GetCity(host)
		err = WriteResult(err, w, "\nLocations\n---------\n")
		code, failed = IsError(w, err)
		if failed {
			return
		}
		isoCode := strings.ToLower(city.Country.IsoCode)
		if _, ok := city.Country.Names[isoCode]; !ok {
			isoCode = "en"
		}
		// WriteResult uses accumulated error
		err = WriteResult(err, w, "Country: %v\n", city.Country.Names[isoCode])
		err = WriteResult(err, w, "City: %v\n", city.City.Names[isoCode])
		err = WriteResult(err, w, "Latitude: %v\n", city.Location.Latitude)
		err = WriteResult(err, w, "Longitude: %v\n", city.Location.Longitude)
		err = WriteResult(err, w, "TimeZone: %v\n", city.Location.TimeZone)
		code, failed = IsError(w, err)
		if failed {
			return
		}
	})
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			loggerInfo.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		loggerInfo.Printf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
	loggerInfo.Println("stopped")
}
