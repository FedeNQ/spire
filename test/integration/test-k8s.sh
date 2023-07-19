#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "${DIR}" || fail-now "Unable to change to script directory"

export SUITES=suites-k8s/*
KINDVERSION=v0.17.0
KUBECTLVERSION=v1.25.3
K8SIMAGE=kindest/node:v1.25.3
./test.sh $1
