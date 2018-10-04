prototypeVmFile = arguments[0]
instanceName = arguments[1]

vm = JSON.parse(read(prototypeVmFile))

delete vm.metadata.ownerReferences
vm.metadata.name = instanceName
vm.spec.running = true
/*vm.spec.template.spec.affinity = {
  "podAffinity": {
    "requiredDuringSchedulingIgnoredDuringExecution": [
      {
        "labelSelector": {
          "matchExpressions": [
            {
              "key": "guest",
              "operator": "In",
              "values": [
                instanceName
              ]
            }
          ]
        },
        "topologyKey": "kubernetes.io/hostname"
      }
    ]
  }
}*/
vm.status = {}

print(JSON.stringify(vm))
