#!/usr/bin/env bash -eou pipefail

docker run --rm -it \
	-h elxir-build \
	-v "${GOPATH}/src":/go/src \
	-v ~/.bashrc:/root/.bashrc \
	-v ~/.gitconfig:/root/.gitconfig \
	gcr.io/elxir-core-infra/service-base-build:latest
