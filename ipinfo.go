// Copyright 2024 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package main implements main method of IPINFO service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/z0rr0/ipinfo/conf"
	"github.com/z0rr0/ipinfo/handle"
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

func main() {
	defer func() {
		if r := recover(); r != nil {
			loggerInfo.Printf("abnormal termination [%v]: \n\t%v\n", Version, r)
		}
	}()
	version := flag.Bool("version", false, "show version")
	config := flag.String("config", Config, "configuration file")
	flag.Parse()

	buildInfo := &handle.BuildInfo{Version: Version, Revision: Revision, BuildDate: BuildDate, GoVersion: GoVersion}
	if *version {
		fmt.Println(buildInfo.String())
		return
	}

	cfg, err := conf.New(*config)
	if err != nil {
		loggerInfo.Fatal(err)
	}

	srv := &http.Server{
		Addr:           cfg.Addr(),
		Handler:        http.DefaultServeMux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20, // 1MB
		ErrorLog:       loggerInfo,
	}
	loggerInfo.Printf("\n%v\nlisten addr: %v\n", buildInfo.String(), srv.Addr)

	handlers := map[string]func(http.ResponseWriter, *conf.IPInfo, *handle.BuildInfo) error{
		"/short":   handle.TextShortHandler,
		"/compact": handle.TextCompactHandler,
		"/json":    handle.JSONHandler,
		"/xml":     handle.XMLHandler,
		"/html":    handle.HTMLHandler,
		"/version": handle.VersionHandler,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start, code := time.Now(), http.StatusOK
		defer func() {
			loggerInfo.Printf("%-5v %v\t%-12v\t%v",
				r.Method, code, time.Since(start), r.RemoteAddr,
			)
		}()

		info, e := cfg.Info(r)
		if e != nil {
			loggerInfo.Println(e)
			http.Error(w, "ERROR", http.StatusInternalServerError)
			return
		}

		url := strings.TrimRight(r.URL.Path, "/ ")
		if h, ok := handlers[url]; ok {
			e = h(w, info, buildInfo)
		} else {
			e = handle.TextHandler(w, r, cfg, info)
		}

		if e != nil {
			loggerInfo.Println(e)
			http.Error(w, "ERROR", http.StatusInternalServerError)
			code = http.StatusInternalServerError
		}
	})
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, os.Signal(syscall.SIGTERM), os.Signal(syscall.SIGQUIT))
		<-sigint

		if e := srv.Shutdown(context.Background()); e != nil {
			loggerInfo.Printf("HTTP server shutdown error: %v", e)
		}
		close(idleConnsClosed)
	}()

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		loggerInfo.Printf("HTTP server ListenAndServe error: %v", err)
	}

	<-idleConnsClosed

	if err = cfg.Close(); err != nil {
		loggerInfo.Printf("cfg close error: %v\n", err)
	}
	loggerInfo.Println("stopped")
}
