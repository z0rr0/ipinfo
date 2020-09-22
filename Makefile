PROJECTNAME=$(shell basename "$(PWD)")

BIN=$(GOPATH)/bin
VERSION=`bash version.sh`
MAIN=github.com/z0rr0/ipinfo
SOURCEDIR=src/$(MAIN)
CONTAINER=container_build.sh
DOCKER_TAG=z0rr0/ipinfo
CONFIG=config.example.json

PID=/tmp/.$(PROJECTNAME).pid
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# MAKEFLAGS += --silent

all: test

build:
	go build -ldflags "$(VERSION)" .

lint: build
	go vet $(MAIN)
	golint $(MAIN)
	go vet $(MAIN)/conf
	golint $(MAIN)/conf

test: lint
	@-cp $(CONFIG) /tmp/
	go test -race -v -cover -coverprofile=conf_coverage.out -trace conf_trace.out $(MAIN)/conf
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	# go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(MAIN)

docker: lint
	bash $(CONTAINER)
	docker build -t $(DOCKER_TAG) .

docker-no-cache: lint
	bash $(CONTAINER)
	docker build --no-cache -t $(DOCKER_TAG) .

start: build
	@echo "  >  $(PROJECTNAME)"
	@-$(BIN)/$(PROJECTNAME) -config config.example.json & echo $$! > $(PID)
	@-cat $(PID)

stop:
	@-touch $(PID)
	@-cat $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

restart: stop start

arm:
	env GOOS=linux GOARCH=arm go install -ldflags "$(VERSION)" $(MAIN)

linux:
	env GOOS=linux GOARCH=amd64 go install -ldflags "$(VERSION)" $(MAIN)

clean: stop
	rm -rf $(BIN)/*
	find $(GOPATH)/$(SOURCEDIR)/ -type f -name "*.out" -print0 -delete
