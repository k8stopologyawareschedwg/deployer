#!/bin/bash

set -eu

VERSION="${1}"
FILES="
deployer-${VERSION}-linux-amd64.gz
deployer-${VERSION}-manifests-allinone.yaml
"

for artifact in $FILES; do
	if [ ! -f "_out/${artifact}" ]; then
		echo "MISSING: ${artifact}" >&2
		exit 1
	fi
done

pushd _out
:> SHA256SUMS
sha256sum ${FILES} >> SHA256SUMS
popd
