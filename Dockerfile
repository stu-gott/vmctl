FROM busybox

RUN wget \
  https://github.com/kubevirt/kubevirt/releases/download/v0.5.0/virtctl-v0.5.0-linux-amd64 \
  && chmod a+x /virtctl*

ADD launchVM /launchVM

ENTRYPOINT /launchVM
