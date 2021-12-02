# Troubleshooting deployer issues

## Openshift

### RTE pods in CrashLoopBackoff, permission denied

#### Symptoms:
something similar in the RTE logs:
```
F1201 08:34:24.836259       1 main.go:98] failed to execute: open /host-run-rte/notify: permission denied
```
**AND** something similar in the RTE events (`oc describe pod -n $RTE_NAMESPACE $RTE_POD_NAME...`)
```
  Warning  FailedCreatePodSandBox  88s   kubelet            Failed to create pod sandbox: rpc error: code = Unknown desc = container create failed: time="2021-12-01T09:02:40Z" level=error msg="container_linux.go:367: starting container process caused: process_linux.go:495: container init caused: failed to set /proc/self/attr/keycreate on procfs: write /proc/self/attr/keycreate: invalid argument"
```

#### Cause:
Obsolete or broken SELinux policy

#### Resolution:
1. undeploy topology-updater components:
```
deployer remove topology-updater
```
2. wait for the relevant Machine Config Pool (MCP) to be updated
3. verify the machineconfig added by the deployer was removed
4. update the deployer to the last stable release
5. deploy again:
```
deployer deploy topology-updater
```
**NOTE** that this will trigger *again* the installation of the most up to date selinux policy. This will take a while and will cause all the worker node to reboot. This is the expected behaviour.
After the nodes rebooted, the daemonset are expected to heal and go running correctly.

If after having waited the Machine Config Pool (MCP) was updated correctly and the node rebooted the daemonset is still not running, please file a [issue](https://github.com/k8stopologyawareschedwg/deployer/issues).
A future version of the deployer will add an option to wait for the MCP to be properly updated before to continue deploy the daemonsets.
