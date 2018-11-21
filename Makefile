TAG=quay.io/fabiand/vmctl

current_dir = $(shell pwd)
output = $(current_dir)/_out

ifdef BUILD_NEXT
    CONTAINER_ENGINE := podman
else
    CONTAINER_ENGINE := docker
endif

build:
	$(CONTAINER_ENGINE) build -t $(TAG) .

run:
	$(CONTAINER_ENGINE) run --rm -it --entrypoint /bin/sh --privileged $(TAG)

build-go: format
	cd cmd/vmctl ;\
	go build vmctl.go

format:
	cd cmd && go fmt ./...
	cd pkg && go fmt ./...

push:
	$(CONTAINER_ENGINE) push $(TAG)

test:
	cd pkg && go test ./...

.PHONY: format docker test
