# IPINFO

![Go](https://github.com/z0rr0/ipinfo/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/ipinfo.svg)
![License](https://img.shields.io/github/license/z0rr0/ipinfo.svg)

IP info web service.

### Build

```bash
make build
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
chmod u+x ipinfo
./ipinfo -config ipinfo.json
```

For docker container

```bash
# mydir/ipinfo.json
# mydir/GeoLite2-City.mmdb
docker run --rm --name ipinfo -u $UID:$UID -p 8082:8082 -v /mydir:/data/conf:ro z0rr0/ipinfo:latest
```

### License

This source code is governed by a [BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause) 
license that can be found in the [LICENSE](https://github.com/z0rr0/ipinfo/blob/master/LICENSE) file.

_This product includes GeoLite2 data created by MaxMind, available from [http://www.maxmind.com](http://www.maxmind.com)_
