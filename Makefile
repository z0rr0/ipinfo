PROGRAM=ipinfo
BIN=bin/ipinfo
VERSION=`bash version.sh`
MAIN=github.com/z0rr0/ipinfo
SOURCEDIR=src/$(MAIN)
CONTAINER=container_build.sh
DOCKER_TAG=z0rr0/ipinfo


all: test

install:
	go install -ldflags "$(VERSION)" $(MAIN)

lint: install
	go vet $(MAIN)
	golint $(MAIN)
	go vet $(MAIN)/db
	golint $(MAIN)/db
	go vet $(MAIN)/conf
	golint $(MAIN)/conf

test: lint
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	# go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(MAIN)

docker: lint
	bash $(CONTAINER)
	docker build -t $(DOCKER_TAG) .

arm:
	env GOOS=linux GOARCH=arm go install -ldflags "$(VERSION)" $(MAIN)

linux:
	env GOOS=linux GOARCH=amd64 go install -ldflags "$(VERSION)" $(MAIN)

clean:
	rm -rf $(GOPATH)/bin/$(PROGRAM) $(GOPATH)/bin/alpine $(GOPATH)/$(SOURCEDIR)/*.out
