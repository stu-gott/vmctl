TAG=quay.io/fabiand/vmctl

build:
	buildah bud -t $(TAG) .

run:
	podman run --rm -it --entrypoint /bin/sh --privileged quay.io/fabiand/vmctl

push:
	
