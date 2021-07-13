# topology-aware-scheduling deployer

`deployer` is a set of go packages and a command line tool to setup all the components and settings needed to enable
the topology-aware-scheduling on a kubernetes cluster. Additionally, the tool can validate if the cluster configuration
is compatible to the topology-aware-scheduling requirements.

## requirements

* kubernetes >= 1.21
* a valid `kubeconfig`
* **validation only** `kubectl` >= 1.21 in your `PATH`

## how does it work?


### validate the cluster configuration:

A kind cluster with the correct configuration:
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: KubeletConfiguration
  cpuManagerPolicy: "static"
  topologyManagerPolicy: "single-numa-node"
  reservedSystemCPUs: "0,16"
  featureGates:
    KubeletPodResourcesGetAllocatable: true
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
```

Does pass the validation:
```
$ ./deployer validate
PASSED>>: the cluster configuration looks ok!
```

A kind cluster configured like this
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: KubeletConfiguration
  cpuManagerPolicy: "static"
  reservedSystemCPUs: "0,16"
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
```

Does **not** pass the validation:
```
$ ./deployer validate
ERROR#000: Incorrect configuration of node "kind-worker" area "kubelet" component "feature gates" setting "": expected "present" detected "missing data"
ERROR#001: Incorrect configuration of node "kind-worker" area "kubelet" component "topology manager" setting "policy": expected "single-numa-node" detected "none"
ERROR#002: Incorrect configuration of node "kind-worker2" area "kubelet" component "feature gates" setting "": expected "present" detected "missing data"
ERROR#003: Incorrect configuration of node "kind-worker2" area "kubelet" component "topology manager" setting "policy": expected "single-numa-node" detected "none"
ERROR#004: Incorrect configuration of node "kind-worker3" area "kubelet" component "feature gates" setting "": expected "present" detected "missing data"
ERROR#005: Incorrect configuration of node "kind-worker3" area "kubelet" component "topology manager" setting "policy": expected "single-numa-node" detected "none"
```

## license
(C) 2021 Red Hat Inc and licensed under the Apache License v2

## build
just run
```bash
make
```

## releases

TBD
