
IMAGEREGISTRY ?= escoffier
IMAGETAG ?= latest

build:
	go mod tidy
	$(GOBUILD) -o ./bin/event-monitor ./

image: build
	docker build -t $(IMAGEREGISTRY)/event-monitor:$(IMAGETAG) -f docker/Dockerfile .