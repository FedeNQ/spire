#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "${DIR}" || fail-now "Unable to change to script directory"

export SUITES=suites-k8s/*
export KUBECTLVERSION=v1.25.3
./test.sh $1
