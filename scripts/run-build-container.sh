#!/usr/bin/env bash -eou pipefail

BUILD_USER="builder"

docker run --rm -it \
	-h elxir-build \
	-u ${BUILD_USER} \
	-v "${GOPATH}/src":/go/src \
	-v ~/.bashrc:"/home/${BUILD_USER}/.bashrc" \
	-v ~/.gitconfig:"/home/${BUILD_USER}/.gitconfig" \
	gcr.io/elxir-core-infra/service-base-build:latest
