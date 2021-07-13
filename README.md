# topology-aware-scheduling deployer

`deployer` is a set of go packages and a command line tool to setup all the components and settings needed to enable
the topology-aware-scheduling on a kubernetes cluster. Additionally, the tool can validate if the cluster configuration
is compatible to the topology-aware-scheduling requirements.

## requirements

* kubernetes >= 1.21
* a valid `kubeconfig`
* **validation only** `kubectl` >= 1.21 in your `PATH`

## license
(C) 2021 Red Hat Inc and licensed under the Apache License v2

## build
just run
```bash
make
```

## releases

TBD
