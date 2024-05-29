
IMAGEREGISTRY ?= escoffier
IMAGETAG ?= latest

PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
NAME = $(shell basename $(PKG))
COMMIT_SHA ?= $(shell git rev-parse --short HEAD)
APPVERSION_KEY := $(PKG)/version.AppVersion
GITVERSION_KEY := $(PKG)/version.GitRevision
LDFLAGS ?= "-X $(APPVERSION_KEY)=$(VERSION) -X $(GITVERSION_KEY)=$(COMMIT_SHA)"
GO = GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go

GOBUILD=$(GO) build -a -ldflags $(LDFLAGS)


build:
	go mod tidy
	$(GOBUILD) -o ./bin/event-monitor ./

image: build
	docker build -t $(IMAGEREGISTRY)/event-monitor:$(IMAGETAG) -f docker/Dockerfile .