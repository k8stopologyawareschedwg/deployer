#!/bin/bash

set -eux

VERSION="${1}"

cp _out/deployer-manifests-allinone.yaml deployer-${VERSION}-manifests-allinone.yaml
cp _out/deployer deployer-${VERSION}-linux-amd64
gzip deployer-${VERSION}-linux-amd64
