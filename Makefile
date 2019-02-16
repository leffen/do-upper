
include .env


APP=do-upper

SOURCES=main.go

VERSION=$(shell git describe --always --long)
BuildTime=$(shell date +%Y-%m-%d.%H:%M:%S)
GitDirty=$(shell git status --porcelain --untracked-files=no)
GitCommit=$(shell git rev-parse --short HEAD)
ifneq ($(GitDirty),)
	GitCommit:= $(GitCommit)-dirty
endif

CGO_ENABLED ?= 0
GOOS ?= linux
GOARCH ?= amd64
GOBUILDFLAGS ?= CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH)

DOCKER_REPO = leffen
DOCKER_NAME = do-upper
DOCKER_IMAGE = $(DOCKER_REPO)/$(DOCKER_NAME)

DOCKER_TAG = v0.1.0
 
compiled: 
	$(GOBUILDFLAGS) go build -ldflags "-s -X main.CommitHash=$(GitCommit) -X main.BuildTime=$(BuildTime)"  -a -installsuffix cgo -o bin/alpine/$(APP) .
	
buildd: test compiled 
	docker build  -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

pushd:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

released: test bump buildd pushd

bump:
	bump_version patch main.go 

build: test
	go build -ldflags  "-s -X main.CommitHash=$(VERSION) -X main.BuildTime=$(BuildTime) -w"  -o bin/$(APP) .

run-serve: 
		go run  $(SOURCES) serve
	
run-status: 
		go run  $(SOURCES) status 

test: vet
	@# this target should always be listed first so "make" runs the tests.
	go test -cover -race -v ./...

vet: 
	@# We can't vet the vendor directory, it fails.
	go list ./... | grep -v vendor | xargs go vet
		