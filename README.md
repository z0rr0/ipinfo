# IPINFO

![Go](https://github.com/z0rr0/ipinfo/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/ipinfo.svg)
![License](https://img.shields.io/github/license/z0rr0/ipinfo.svg)

IP info web service.

### Build

```bash
make build
# cp config.example.json ipinfo.json
# update ipinfo.json
./ipinfo -config ipinfo.json
```

For docker container `z0rr0/ipinfo:latest`

```bash
make docker
# or only for linux amd64
# make docker_linux_amd64
```

### Local run

```bash
make start
make stop

# or alias for [stop + start]
make restart
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
