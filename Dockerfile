FROM fedora

RUN curl -LO \
  https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
  && chmod a+x /kubectl

RUN curl -LO \
  https://github.com/kubevirt/kubevirt/releases/download/v0.5.0/virtctl-v0.5.0-linux-amd64 \
  && chmod a+x /virtctl*

ADD launchVM /launchVM

ENTRYPOINT ["/launchVM"]
