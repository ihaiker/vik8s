#!/bin/bash

HOSTNAME=registry.terraform.io
NAMESPACE=xhaiker
NAME=vik8s
BINARY=bin/terraform-provider-${NAME}
VERSION=$(git describe --tags `git rev-list --tags --max-count=1`)
VERSION=${VERSION:1}
OS_ARCH=darwin_amd64

go build -o ${BINARY} cmd/plugins.go

mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
