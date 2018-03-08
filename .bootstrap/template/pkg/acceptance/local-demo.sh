#!/usr/bin/env bash

set -eou pipefail
#set -x  # useful for debugging

docker_cleanup() {
    echo "cleaning up existing network and containers..."
    CONTAINERS='servicename'
    docker ps | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker stop {} || true
    docker ps -a | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker rm {} || true
    docker network list | grep ${CONTAINERS} | awk '{print $2}' | xargs -I {} docker network rm {} || true
}

# optional settings (generally defaults should be fine, but sometimes useful for debugging)
SERVICENAME_LOG_LEVEL="${SERVICENAME_LOG_LEVEL:-INFO}"  # or DEBUG
SERVICENAME_TIMEOUT="${SERVICENAME_TIMEOUT:-5}"  # 10, or 20 for really sketchy network

# local and filesystem constants
LOCAL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# container command constants
SERVICENAME_IMAGE="gcr.io/elxir-core-infra/servicename:snapshot" # develop

echo
echo "cleaning up from previous runs..."
docker_cleanup

echo
echo "creating servicename docker network..."
docker network create servicename

# TODO start and healthcheck dependency services if necessary

echo
echo "starting servicename..."
port=10100
name="servicename-0"
docker run --name "${name}" --net=servicename -d -p ${port}:${port} ${SERVICENAME_IMAGE} \
    start \
    --logLevel "${SERVICENAME_LOG_LEVEL}" \
    --serverPort ${port}
    # TODO add other relevant args if necessary
servicename_addrs="${name}:${port}"
servicename_containers="${name}"

echo
echo "testing servicename health..."
docker run --rm --net=servicename ${SERVICENAME_IMAGE} test health \
    --servicenames "${servicename_addrs}" \
    --logLevel "${SERVICENAME_LOG_LEVEL}"

echo
echo "testing servicename ..."
docker run --rm --net=servicename ${SERVICENAME_IMAGE} test io \
    --servicenames "${servicename_addrs}" \
    --logLevel "${SERVICENAME_LOG_LEVEL}"
    # TODO add other relevant args if necessary

echo
echo "cleaning up..."
docker_cleanup

echo
echo "All tests passed."
