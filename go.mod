module github.com/z0rr0/ipinfo

go 1.26

toolchain go1.26.2

require (
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/oschwald/geoip2-golang v1.13.0
)

require (
	github.com/4meepo/tagalign v1.4.3 // indirect
	github.com/alfatraining/structtag v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/oschwald/maxminddb-golang v1.13.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
)

tool (
	github.com/4meepo/tagalign/cmd/tagalign
	golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
)
