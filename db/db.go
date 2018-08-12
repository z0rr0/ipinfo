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
	"net/http"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
)

func link(url, fileName string) string {
	return fmt.Sprintf("%v/%v", url, fileName)
}

func download(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fd, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fd, resp.Body)
	return err
}

func checkMD5(url, sumFile, fileName, dbFile string) error {
	fd, err := os.Open(dbFile)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	if _, err := io.Copy(h, fd); err != nil {
		return err
	}
	currentMD5 := hex.EncodeToString(h.Sum(nil))

	resp, err := http.Get(link(url, sumFile))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed md5 sum response %v", resp.Status)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if newMD5 := string(bodyBytes); currentMD5 == newMD5 {
		// md5 sum didn't change
		return nil
	}
	// md5 sum is differ, download new db file
	// ignore error
	os.Remove(fileName)
	return download(link(url, fileName), fileName)
}

// GetDb downloads and save IP database.
func GetDb(url, file, checkSum string) (*geoip2.Reader, error) {
	tmp := os.TempDir()
	fileName := filepath.Join(tmp, file)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// download 1st time
		err = download(link(url, file), fileName)
		if err != nil {
			return nil, err
		}
	} else {
		// checksum
		err = checkMD5(url, checkSum, fileName, file)
		if err != nil {
			return nil, err
		}
	}
	return geoip2.Open(fileName)
}
