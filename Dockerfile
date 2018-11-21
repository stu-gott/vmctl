FROM fedora AS build

RUN yum install -y golang make
ENV GOPATH=/go
RUN mkdir -p /go/src/kubevirt.io/vmctl/cmd/vmctl
RUN mkdir -p /go/src/kubevirt.io/vendor
COPY . /go/src/kubevirt.io/vmctl/

WORKDIR /go/src/kubevirt.io/vmctl
RUN make build-go

FROM build AS test

WORKDIR /go/src/kubevirt.io/vmctl/pkg
RUN go test -cover -v -race ./...

FROM fedora

COPY --from=build /go/src/kubevirt.io/vmctl/cmd/vmctl/vmctl /vmctl

ENTRYPOINT ["/vmctl"]
