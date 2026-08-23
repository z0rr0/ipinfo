TARGET=ipinfo
TS=$(shell date -u +"%FT%T")
TAG=$(shell git tag | sort -V | tail -1 | grep . || echo "v0.0.0")
COMMIT=$(shell git log --oneline | head -1)
VERSION=$(firstword $(COMMIT))

# GO_LDFLAGS, not LDFLAGS: the latter conventionally carries C linker flags and
# is commonly exported by the host shell (Homebrew sets it), which would leak
# into the build.
GO_LDFLAGS=-X main.Version=$(TAG) -X main.Revision=git:$(VERSION) -X main.BuildDate=$(TS)

IMAGE=z0rr0/ipinfo
# Docker Hub tags of this image carry no "v" prefix: 1.10.0, 1.9.11, ...
IMAGE_TAG=$(TAG:v%=%)
BUILDER=ipinfo-multiarch
DOCKER_PLATFORMS=linux/amd64,linux/arm64

# local config for development and testing
#CONFIG=local/config.json
CONFIG=config.example.json
TEST_CONFIG=/tmp/ipinfo_test.json
TEST_STORAGE=/tmp/GeoLite2-City.mmdb
URL_STORAGE=https://static.fwtf.xyz/other/GeoLite2-City.mmdb

PID=/tmp/.$(TARGET).pid
STDERR=/tmp/.$(TARGET)-stderr.txt

all: test

build: lint
	go build -o $(PWD)/$(TARGET) -ldflags "$(GO_LDFLAGS)"

fmt:
	gofmt -d .

check_fmt:
	@test -z "`gofmt -l .`" || { echo "ERROR: failed gofmt, for more details run - make fmt"; false; }
	@-echo "gofmt successful"

lint: check_fmt
	go vet $(PWD)/...
	-golangci-lint -c golangci.yml run $(PWD)/...
	-govulncheck ./...
	-staticcheck ./...
	-gosec ./...

gh: check_fmt prepare
	go vet $(PWD)/...
	go test -race -cover $(PWD)/...

prepare:
	@-cp -f $(CONFIG) $(TEST_CONFIG)
	@test -f $(TEST_STORAGE) || curl -o $(TEST_STORAGE) $(URL_STORAGE)

test: lint prepare
	# go test -v -race -cover -coverprofile=coverage.out -trace trace.out github.com/z0rr0/ipinfo
	# go tool cover -html=coverage.out
	go test -race -cover $(PWD)/...

# Host-arch only: the default Docker Desktop builder uses the overlay2 image
# store, which cannot --load a multi-arch manifest list.
docker: lint
	docker buildx build \
		-t $(IMAGE):latest -t $(IMAGE):$(IMAGE_TAG) \
		--build-arg GO_LDFLAGS="$(GO_LDFLAGS)" \
		--load .

# Multi-arch goes straight to the registry: the `docker` driver cannot build a
# manifest list at all, so this needs its own docker-container builder, created
# on first use and reused afterwards. Not a dependency of `docker` — that would
# build the image twice.
docker-push: lint
	@test "$(TAG)" != "v0.0.0" || (echo "no git tag found, refusing to push $(IMAGE):v0.0.0" && exit 1)
	@docker buildx inspect $(BUILDER) >/dev/null 2>&1 || docker buildx create --name $(BUILDER) --driver docker-container
	docker buildx build --builder $(BUILDER) \
		--platform $(DOCKER_PLATFORMS) \
		-t $(IMAGE):latest -t $(IMAGE):$(IMAGE_TAG) \
		--build-arg GO_LDFLAGS="$(GO_LDFLAGS)" \
		--push .

clean:
	rm -f $(PWD)/$(TARGET) $(TEST_CONFIG) $(TEST_STORAGE)
	find ./ -type f -name "*.out" -delete

tools:
	@go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1
	@go get -tool github.com/4meepo/tagalign/cmd/tagalign@latest
	@go get -tool golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	@go get -tool github.com/securego/gosec/v2/cmd/gosec@latest
	@go get -tool honnef.co/go/tools/cmd/staticcheck@latest
	@go get -tool golang.org/x/vuln/cmd/govulncheck@latest

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
