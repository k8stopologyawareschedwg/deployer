#!/bin/bash

set -eux

RESERVED_CPUS=${RESERVED_CPUS:-0,8}
TM_POLICY=${TM_POLICY:-single-numa-node}
TM_SCOPE=${TM_SCOPE:-container}

minikube start \
	--nodes=3 \
	--kvm-numa-count=2 \
	--cpus=16 \
	--extra-config=kubelet.cpu-manager-policy="static" \
	--extra-config=kubelet.reserved-cpus="${RESERVED_CPUS}" \
	--extra-config=kubelet.topology-manager-policy="${TM_POLICY}" \
	--extra-config=kubelet.topology-manager-scope="${TM_SCOPE}"
kubectl label node minikube-m02 node-role.kubernetes.io/worker=''
kubectl label node minikube-m03 node-role.kubernetes.io/worker=''

echo <<< EOF > rte-minikube.yaml
kubelet:
  topologyManagerPolicy: ${TM_POLICY}
  topologyManagerScope: ${TM_SCOPE}
EOF

echo "deployer deploy api"
echo "deployer deploy topology-updater --rte-config-file ./rte-minikube.yaml"
