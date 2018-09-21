prototypeVmFile = arguments[0]
instanceName = arguments[1]

vmInstance = JSON.parse(read(prototypeVmFile))

delete vmInstance.metadata.ownerReferences
vmInstance.metadata.name = instanceName
vmInstance.spec.running = true
/*vmInstance.spec.template.spec.affinity = {
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
vmInstance.status = {}

print(JSON.stringify(vmInstance))
