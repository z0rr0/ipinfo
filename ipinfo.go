// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
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
	Config = "config.json"
	// interruptPrefix is constant prefix of interrupt signal
	interruptPrefix = "interrupt signal"
	timeout         = 30 * time.Second
)

var (
	// Version is LUSS version
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
	defer cfg.Close()

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
		start, code := time.Now(), http.StatusOK
		defer func() {
			loggerInfo.Printf("%-5v %v\t%-12v\t%v",
				r.Method,
				code,
				time.Since(start),
				r.RemoteAddr,
			)
		}()
		host, err := cfg.GetIP(r)
		if err != nil {
			loggerInfo.Println(err)
			code = http.StatusInternalServerError
			http.Error(w, "ERROR", code)
			return
		}
		// main info
		fmt.Fprintf(w, "IP: %v\nProto: %v\nMethod: %v\nURI: %v\n",
			host, r.Proto, r.Method, r.RequestURI)

		fmt.Fprintln(w, "\nHeaders\n---------")
		for k, v := range r.Header {
			if !cfg.IsIgnoredHeader(k) {
				fmt.Fprintf(w, "%v: %v\n", k, strings.Join(v, "; "))
			}
		}
		// init form load
		r.FormValue("test")
		fmt.Fprintln(w, "\nParams\n---------")
		for k, v := range r.Form {
			fmt.Fprintf(w, "%v: %v\n", k, strings.Join(v, "; "))
		}
		// additional info
		city, err := cfg.GetCity(host)
		if err == nil {
			fmt.Fprintln(w, "\nLocations\n---------")
			isoCode := strings.ToLower(city.Country.IsoCode)
			if _, ok := city.Country.Names[isoCode]; !ok {
				isoCode = "en"
			}
			fmt.Fprintf(w, "Country: %v\n", city.Country.Names[isoCode])
			fmt.Fprintf(w, "City: %v\n", city.City.Names[isoCode])
			fmt.Fprintf(w, "Latitude: %v\n", city.Location.Latitude)
			fmt.Fprintf(w, "Longitude: %v\n", city.Location.Longitude)
			fmt.Fprintf(w, "TimeZone: %v\n", city.Location.TimeZone)
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
