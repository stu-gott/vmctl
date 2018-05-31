TAG=quay.io/fabiand/vmctl

build:
	buildah bud -t $(TAG) .
