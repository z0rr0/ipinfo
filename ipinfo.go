// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package main implements main method of IPINFO service.
package main

import (
	"runtime"
	"log"
	"os"
	"fmt"
	"flag"
)

const (
	// Name is a program name
	Name = "IPINFO"
	// Config is default configuration file name
	Config = "config.json"
	// interruptPrefix is constant prefix of interrupt signal
	interruptPrefix = "interrupt signal"
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
	//config := flag.String("config", Config, "configuration file")
	flag.Parse()

	if *version {
		fmt.Printf("\tVersion: %v\n\tRevision: %v\n\tBuild date: %v\n\tGo version: %v\n",
			Version, Revision, BuildDate, GoVersion)
		return
	}


}
