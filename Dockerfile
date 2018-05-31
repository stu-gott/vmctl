FROM kubevirt/libvirtd
ADD launchVM /launchVM
ENTRYPOINT /launchVM
