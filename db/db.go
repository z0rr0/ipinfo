// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package db contains methods and structures for MaxMind GeoLite2 database.
package db

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"compress/gzip"
	"github.com/oschwald/geoip2-golang"
)

var (
	// internal logger
	loggerInfo = log.New(os.Stdout, "INFO [db]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func extract(archivedFile, filePath string) error {
	loggerInfo.Printf("extract %v\n", archivedFile)
	if archivedFile == "" {
		return fmt.Errorf("empty file name")
	}

	gzFd, err := os.Open(archivedFile)
	if err != nil {
		return err
	}
	defer gzFd.Close()

	fd, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	f, err := gzip.NewReader(gzFd)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(fd, f)
	if err != nil {
		return err
	}
	loggerInfo.Printf("extracted file %v\n", filePath)
	return nil
}

func link(url, fileName, format string) string {
	return fmt.Sprintf("%v/%v%v", url, fileName, format)
}

func download(url, fileName, format string) error {
	loggerInfo.Printf("download db: %v\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	archivedFile := fileName + format
	fd, err := os.Create(archivedFile)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return err
	}
	return extract(archivedFile, fileName)
}

func checkMD5(url, sumFile, dbFile, fileName, format string) error {
	loggerInfo.Printf("check md5sum: %v\n", fileName)
	fd, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	if _, err := io.Copy(h, fd); err != nil {
		return err
	}
	currentMD5 := hex.EncodeToString(h.Sum(nil))

	resp, err := http.Get(link(url, sumFile, ""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed md5 sum response %v", resp.Status)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	newMD5 := string(bodyBytes)
	if currentMD5 == newMD5 {
		// md5 sum didn't change
		return nil
	}
	loggerInfo.Printf("diff md5sum %v!=%v\n", currentMD5, newMD5)
	// md5 sum is differ, download new db file
	// ignore error
	os.Remove(fileName)
	return download(link(url, dbFile, format), fileName, format)
}

// GetDb downloads and save IP database.
func GetDb(url, file, checkSum, format string) (*geoip2.Reader, error) {
	tmp := os.TempDir()
	fileName := filepath.Join(tmp, file)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// download 1st time
		err = download(link(url, file, format), fileName, format)
		if err != nil {
			return nil, err
		}
	} else {
		// checksum
		err = checkMD5(url, checkSum, file, fileName, format)
		if err != nil {
			return nil, err
		}
	}
	return geoip2.Open(fileName)
}
