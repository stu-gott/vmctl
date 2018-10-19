TAG=quay.io/fabiand/vmctl

ifdef BUILD_NEXT
build: build-go
	buildah bud -t $(TAG) .

run:
	podman run --rm -it --entrypoint /bin/sh --privileged quay.io/fabiand/vmctl

push:
	echo
endif

build: build-go
	docker build -t $(TAG) cmd/vmctl/

build-go: format
	cd cmd/vmctl ;\
	go build vmctl.go

format:
	go fmt cmd/vmctl/vmctl.go

.PHONY: format docker
