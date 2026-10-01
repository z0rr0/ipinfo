// Copyright 2025 Aleksandr Zaitsev <me@axv.email>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package main implements main method of IPINFO service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
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
	// Name is a program name.
	Name = "IPINFO"
	// Config is default configuration file name.
	Config  = "config.json"
	timeout = 30 * time.Second
)

var (
	// Version is program git version.
	Version = "" //nolint:gochecknoglobals
	// Revision is revision number.
	Revision = "" //nolint:gochecknoglobals
	// BuildDate is build date.
	BuildDate = "" //nolint:gochecknoglobals
	// GoVersion is runtime Go language version.
	GoVersion = runtime.Version() //nolint:gochecknoglobals

	// internal logger.
	loggerInfo = log.New(os.Stdout, fmt.Sprintf("INFO [%v]: ", Name), log.Ldate|log.Ltime|log.Lshortfile) //nolint:gochecknoglobals
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
		fmt.Println(buildInfo.String()) //nolint:forbidigo
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
	initLogger(true, os.Stdout)
	loggerInfo.Printf("\n%v\nlisten addr: %v\n", buildInfo.String(), srv.Addr)

	http.Handle("/", newHandler(cfg, buildInfo))

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

// newHandler returns the HTTP handler that routes requests by the trimmed URL path.
// "/health" is served before the client IP lookup and is not written to the access log.
func newHandler(cfg *conf.Cfg, buildInfo *handle.BuildInfo) http.Handler {
	handlers := map[string]func(http.ResponseWriter, *conf.IPInfo, *handle.BuildInfo) error{
		"/short":   handle.TextShortHandler,
		"/compact": handle.TextCompactHandler,
		"/ip":      handle.IPHandler,
		"/json":    handle.JSONHandler,
		"/xml":     handle.XMLHandler,
		"/html":    handle.HTMLHandler,
		"/full":    handle.FullHTMLHandler,
		"/version": handle.VersionHandler,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := strings.TrimRight(r.URL.Path, "/ ")
		if url == "/health" {
			if e := handle.HealthHandler(w); e != nil {
				loggerInfo.Println(e)
			}
			return
		}

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
}

// initLogger initializes logger with debug mode and writer.
func initLogger(debug bool, w io.Writer) {
	var level = slog.LevelInfo

	if debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})))
}
