// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package db contains methods and structures for MaxMind GeoLite2 database.
package db

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
)

var (
	// internal logger
	loggerInfo = log.New(os.Stdout, "INFO [db]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

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

	// ungzip
	fz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer fz.Close()

	_, err = io.Copy(fd, fz)
	if err != nil {
		return err
	}
	loggerInfo.Printf("file %v downloaded and extracted\n", fileName)
	return nil
}

func checkMD5(url, dbFielURL, fileName string) error {
	loggerInfo.Printf("check md5sum of %v\n", fileName)
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

	resp, err := http.Get(url)
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
		loggerInfo.Println("md5sum is same")
		return nil
	}
	loggerInfo.Printf("diff md5sum %v!=%v\n", currentMD5, newMD5)
	os.Remove(fileName) // ignore error
	return download(dbFielURL, fileName)
}

// GetDb downloads and saves IP database.
func GetDb(url, file, checkSum, format string) (*geoip2.Reader, error) {
	tmp := os.TempDir()
	localFileName := filepath.Join(tmp, file)
	dbFileURL := fmt.Sprintf("%v/%v%v", url, file, format)
	checksumURL := fmt.Sprintf("%v/%v", url, checkSum)

	loggerInfo.Printf("get db:\n\tlocal file: %v\n\tdownload link: %v\n\tchecksum link: %v\n",
		localFileName, dbFileURL, checksumURL)

	if _, err := os.Stat(localFileName); os.IsNotExist(err) {
		// download 1st time
		err = download(dbFileURL, localFileName)
		if err != nil {
			return nil, err
		}
	} else {
		// checksum
		err = checkMD5(checksumURL, dbFileURL, localFileName)
		if err != nil {
			return nil, err
		}
	}
	return geoip2.Open(localFileName)
}
