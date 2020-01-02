// Copyright 2020 Alexander Zaytsev <thebestzorro@yandex.ru>.
// All rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the LICENSE file.

// Package db contains methods and structures for MaxMind GeoLite2 database.
package db

import (
	"fmt"
	"os"

	"github.com/oschwald/geoip2-golang"
)

// GetDb downloads and saves IP database.
func GetDb(fileName string) (*geoip2.Reader, error) {
	_, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("db file %v not found: %v", fileName, err)
	}
	return geoip2.Open(fileName)
}
