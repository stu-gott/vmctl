FROM fedora

RUN curl -LO \
  https://storage.googleapis.com/kubernetes-release/release/v1.11.3/bin/linux/amd64/kubectl \
  && chmod a+x kubectl
ENV KUBECTL=/kubectl

RUN curl -L -o virtctl \
  https://github.com/kubevirt/kubevirt/releases/download/v0.8.0/virtctl-v0.8.0-linux-amd64 \
  && chmod a+x virtctl
ENV VIRTCTL=/virtctl

RUN dnf install -y js

ADD . /vmctl

ENTRYPOINT ["/vmctl/launchVM"]
