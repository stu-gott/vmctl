prototypeVmFile = arguments[0]
instanceName = arguments[1]
nodeName = arguments[2]

vm = JSON.parse(read(prototypeVmFile))

delete vm.metadata.ownerReferences
vm.metadata.name = instanceName
vm.spec.running = true
vm.spec.template.spec.nodeSelector = {"kubernetes.io/hostname": nodeName}
vm.status = {}

print(JSON.stringify(vm))
