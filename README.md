# IPINFO

[![GoDoc](https://godoc.org/github.com/z0rr0/ipinfo?status.svg)](https://godoc.org/github.com/z0rr0/ipinfo)

IP info web service.


### Build

```bash
make install
```

For docker container

```bash
make docker 
```

### Run

```bash
# start
make start

# stop
make stop

# restart
make restart

# run with custom config
$GOPATH/bin/ipinfo -config ipinfo.json
```

For docker container

```bash
docker run --rm --name ipinfo -p 8080:8080 -v /mydir:/data/conf:ro ipinfo
```

### License

This source code is governed by a [BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause) license that can be found in the [LICENSE](https://github.com/z0rr0/ipinfo/blob/master/LICENSE) file.


_This product includes GeoLite2 data created by MaxMind, available from [http://www.maxmind.com](http://www.maxmind.com)_
