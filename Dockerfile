FROM fedora AS builder

RUN yum install -y golang make
ENV GOPATH=/go
RUN mkdir -p /go/src/kubevirt.io/vmctl/cmd/vmctl
RUN mkdir -p /go/src/kubevirt.io/vendor
COPY cmd/vmctl/vmctl.go /go/src/kubevirt.io/vmctl/cmd/vmctl/vmctl.go
COPY vendor /go/src/kubevirt.io/vmctl/vendor/

WORKDIR /go/src/kubevirt.io/vmctl/cmd/vmctl/
RUN go build vmctl.go

FROM fedora

COPY --from=builder /go/src/kubevirt.io/vmctl/cmd/vmctl/vmctl /vmctl

ENTRYPOINT ["/vmctl"]
