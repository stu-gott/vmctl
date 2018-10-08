TAG=quay.io/fabiand/vmctl

ifdef BUILD_NEXT
docker: build
	buildah bud -t $(TAG) .

run:
	podman run --rm -it --entrypoint /bin/sh --privileged quay.io/fabiand/vmctl

push:
	echo
endif

docker: build
	docker build -t $(TAG) cmd/vmctl/

build: format
	pushd cmd/vmctl > /dev/null ;\
	go build vmctl.go ;\
	popd > /dev/null

format:
	go fmt ./...

.PHONY: format docker
