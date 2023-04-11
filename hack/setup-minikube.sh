#!/bin/bash

set -eu

RESERVED_CPUS=${RESERVED_CPUS:-0,8}
TM_POLICY=${TM_POLICY:-single-numa-node}
TM_SCOPE=${TM_SCOPE:-container}

echo "#>>> starting the minikube cluster"
minikube start \
	--nodes=3 \
	--kvm-numa-count=2 \
	--cpus=16
for node in minikube-m02 minikube-m03; do
	kubectl label node $node node-role.kubernetes.io/worker=''
done

echo "#>>> preparing the configuration extras"
cat << EOF > rte-minikube.yaml
kubelet:
  topologyManagerPolicy: ${TM_POLICY}
  topologyManagerScope: ${TM_SCOPE}
EOF

cat << EOF > kubeletconf-patch.yaml
cpuManagerPolicy: static
cpuManagerPolicyOptions:
  full-pcpus-only: "false"
cpuManagerReconcilePeriod: 5s
memoryManagerPolicy: Static
topologyManagerPolicy: ${TM_POLICY}
topologyManagerScope: ${TM_SCOPE}
evictionHard:
  memory.available: 100Mi
kubeReserved:
  memory: 500Mi
reservedSystemCPUs: ${RESERVED_CPUS}
reservedMemory:
  - numaNode: 0
    limits:
      memory: 600Mi
EOF

cat << EOF > fix-kubeletconf.sh
#!/bin/bash
systemctl stop kubelet
cat /etc/kubeletconf-patch.yaml >> /var/lib/kubelet/config.yaml
rm -f /var/lib/kubelet/cpu_manager_state
rm -f /var/lib/kubelet/memory_manager_state
systemctl start kubelet
EOF

echo "#>>> fixing the worker nodes"
for node in minikube-m02 minikube-m03; do
minikube cp \
	kubeletconf-patch.yaml \
	$node:/etc/kubeletconf-patch.yaml
minikube cp \
	fix-kubeletconf.sh \
	$node:/bin/fix-kubeletconf.sh
done

for node in minikube-m02 minikube-m03; do
minikube ssh \
	-n $node \
	sudo bash /bin/fix-kubeletconf.sh
done

# to emphasize the reserve cache behavior:
#     deployer deploy topology-updater --rte-config-file ./rte-minikube.yaml  --updater-notif-enable=false --updater-sync-period=15s

echo "#>>> setup done! now run:"
echo "deployer deploy api"
echo "deployer deploy topology-updater --rte-config-file ./rte-minikube.yaml"
echo "deployer deploy scheduler-plugin --sched-resync-period 5s"
