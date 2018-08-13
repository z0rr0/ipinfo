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
	"net"
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

	if *version {
		fmt.Printf("\tVersion: %v\n\tRevision: %v\n\tBuild date: %v\n\tGo version: %v\n",
			Version, Revision, BuildDate, GoVersion)
		return
	}

	cfg, err := conf.New(*config)
	if err != nil {
		panic(err)
	}
	defer cfg.Storage.Close()

	srv := &http.Server{
		Addr:           cfg.Addr(),
		Handler:        http.DefaultServeMux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20, // 1MB
		ErrorLog:       loggerInfo,
	}
	loggerInfo.Printf("listen addr: %v\n", srv.Addr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loggerInfo.Printf("request %v\n", r.RemoteAddr)
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Fprintln(w, "ERROR")
			return
		}
		// main info
		fmt.Fprintf(w, "IP: %v\nProto: %v\nMethod: %v\nURI: %v\n",
			host, r.Proto, r.Method, r.RequestURI)

		fmt.Fprintln(w, "\nHeaders\n---------")
		for k, v := range r.Header {
			fmt.Fprintf(w, "%v: %v\n", k, strings.Join(v, "; "))
		}

		r.FormValue("test")
		fmt.Fprintln(w, "\nParams\n---------")
		for k, v := range r.Form {
			fmt.Fprintf(w, "%v: %v\n", k, strings.Join(v, "; "))
		}
		// additional info
		ip := net.ParseIP(host)
		city, err := cfg.Storage.City(ip)
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
