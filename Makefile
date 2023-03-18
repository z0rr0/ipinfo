TARGET=ipinfo
TS=$(shell date -u +"%FT%T")
TAG=$(shell git tag | sort -V | tail -1)
COMMIT=$(shell git log --oneline | head -1)
VERSION=$(firstword $(COMMIT))

LDFLAGS=-X main.Version=$(TAG) -X main.Revision=git:$(VERSION) -X main.BuildDate=$(TS)
DOCKER_TAG=z0rr0/ipinfo

CONFIG=config.example.json
TEST_CONFIG=/tmp/ipinfo_test.json
TEST_STORAGE=/tmp/GeoLite2-City.mmdb
URL_STORAGE=https://static.fwtf.xyz/other/GeoLite2-City.mmdb

PID=/tmp/.$(TARGET).pid
STDERR=/tmp/.$(TARGET)-stderr.txt

all: test

build: lint
	go build -o $(PWD)/$(TARGET) -ldflags "$(LDFLAGS)"

fmt:
	gofmt -d .

check_fmt:
	@test -z "`gofmt -l .`" || { echo "ERROR: failed gofmt, for more details run - make fmt"; false; }
	@-echo "gofmt successful"

lint: check_fmt
	go vet $(PWD)/...
	#golint -set_exit_status $(PWD)/...
	#golangci-lint run $(PWD)/...

prepare:
	@-cp -f $(CONFIG) $(TEST_CONFIG)
	@test -f $(TEST_STORAGE) || curl -o $(TEST_STORAGE) $(URL_STORAGE)

test: lint prepare
	# go test -v -race -cover -coverprofile=coverage.out -trace trace.out github.com/z0rr0/ipinfo
	# go tool cover -html=coverage.out
	go test -race -cover $(PWD)/...

docker: lint clean
	docker build --build-arg LDFLAGS="$(LDFLAGS)" -t $(DOCKER_TAG) .

docker_linux_amd64: lint clean
	docker buildx build --platform linux/amd64 --build-arg LDFLAGS="$(LDFLAGS)" -t $(DOCKER_TAG) .

clean:
	rm -f $(PWD)/$(TARGET)
	find ./ -type f -name "*.out" -delete

start: build
	@echo "  >  $(TARGET)"
	@-$(PWD)/$(TARGET) -config $(CONFIG) & echo $$! > $(PID)
	@-cat $(PID)

stop:
	@-touch $(PID)
	@-cat $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

restart: stop start
