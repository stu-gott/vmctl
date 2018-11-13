TAG=quay.io/fabiand/vmctl

current_dir = $(shell pwd)
output = $(current_dir)/_out

ifdef BUILD_NEXT
build:
	buildah bud -t $(TAG) .

run:
	podman run --rm -it --entrypoint /bin/sh --privileged quay.io/fabiand/vmctl

push:
	echo

test:
	rm -rf $(output) && mkdir -p $(output)
	podman build -f Dockerfile.unit_test -v $(output):/output -t test .
endif

build:
	docker build -t $(TAG) .

build-go: format
	cd cmd/vmctl ;\
	go build vmctl.go

format:
	cd cmd && go fmt ./...
	cd pkg && go fmt ./...

test:
	rm -rf $(output) && mkdir -p $(output)
	docker build -f Dockerfile.unit_test -t test .

.PHONY: format docker
