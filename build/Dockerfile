# use stretch (debian) b/c `go test -race` requires glibc, which isn't in the alpine variant
FROM golang:1.9.1-stretch

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev cmake zip unzip tar gzip \
    vim emacs bash-completion locales curl lsb-release openssh-client git

# set lang as UTF-8
RUN rm -rf /var/lib/apt/lists/* && \
    localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG en_US.utf8

# install Docker
ENV DOCKER_VERSION '17.03.0-ce'
RUN wget https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION}.tgz -O /tmp/docker-${DOCKER_VERSION}.tgz && \
    tar xzf /tmp/docker-${DOCKER_VERSION}.tgz -C /tmp && \
    mv /tmp/docker/* /usr/bin && \
    rm /tmp/docker-${DOCKER_VERSION}.tgz && \
    rm -r /tmp/docker

# install Google stuffs
ENV CLOUD_SDK_VERSION 183.0.0
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt ${CLOUD_SDK_REPO} main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update && apt-get install -y \
    google-cloud-sdk=${CLOUD_SDK_VERSION}-0 \
    google-cloud-sdk-datastore-emulator=${CLOUD_SDK_VERSION}-0

ENV GOPATH "/go"

ENTRYPOINT ["/bin/bash"]