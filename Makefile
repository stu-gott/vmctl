TAG=quay.io/fabiand/vmctl

build:
	docker build -t $(TAG) .

push:
	docker push $(TAG)
