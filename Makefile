IMPORT_PATH := github.com/tomsolem/strava

SOURCES := $(shell find . -name "*.go")
OUTPUT := main

V := 1 # When V is set, print commands and build progress.

export GO111MODULE=on

VERSION          := $(shell git describe --tags --always --dirty="-dev")
DATE             := $(shell date -u '+%Y-%m-%d-%H%M UTC')
VERSION_FLAGS    := -ldflags='-X "main.Version=$(VERSION)" -X "main.BuildTime=$(DATE)"'

.PHONY: all
all: clean build

build: $(OUTPUT)

$(OUTPUT): $(SOURCES)
	$Q CGO_ENABLED=0 GOOS=linux go build -a --installsuffix dist -o $(OUTPUT) $(if $V,-v) $(VERSION_FLAGS) $(IMPORT_PATH)/cmd/main

test:
	mkdir -p .cover
	go test $$(go list ./... | grep -v /mock | grep -v /config | grep -v /main | grep -v /v2/service ) -cover -covermode=count -coverprofile=.cover/coverage-all.out

cover: test
	go tool cover -html=.cover/coverage-all.out

clean:
	-rm $(OUTPUT)

run:
	./$(OUTPUT)

.PHONY: format
format:
	#go fmt main.go
	go fmt ./...

setup:
	go mod download
