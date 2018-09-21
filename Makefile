TAG=quay.io/fabiand/vmctl

ifdef BUILD_NEXT
build:
	buildah bud -t $(TAG) .

run:
	podman run --rm -it --entrypoint /bin/sh --privileged quay.io/fabiand/vmctl

push:
	echo
endif

build:
	docker build -t $(TAG) cmd/vmctl/
