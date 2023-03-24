#!/usr/bin/env bash

JQ="/usr/bin/jq"
LISTING_DIR="${1}"

bundle=$( ${JQ} -r '.bundle' /dev/stdin 2>&1 )
pod_ident=$( ${JQ} '"___" + .annotations["io.kubernetes.pod.namespace"] + "___" + .annotations["io.kubernetes.pod.name"]' < ${bundle}/config.json )

touch "${LISTING_DIR}/${pod_ident}" || :
