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
ENV BUILD_USER "builder"
RUN useradd -ms /bin/bash ${BUILD_USER} && \
    chown -R ${BUILD_USER} /usr/local

# install Postgres
ENV POSTGRES_VERSION 10
RUN export POSTGRES_REPO="stretch-pgdg" && \
    echo "deb http://apt.postgresql.org/pub/repos/apt ${POSTGRES_REPO} main" | tee -a /etc/apt/sources.list.d/pgdg.list && \
    curl https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && apt-get install -y \
    postgresql-${POSTGRES_VERSION} \
    postgresql-contrib-${POSTGRES_VERSION} && \
    chown -R ${BUILD_USER} /var/lib/postgresql && \
    chown -R ${BUILD_USER} /run/postgresql && \
    chown -R ${BUILD_USER} /var/log/postgresql && \
    rm /var/log/postgresql/* && \
    rm -r /var/lib/postgresql/${POSTGRES_VERSION}/*

USER ${BUILD_USER}

# init Postgres tests server
ENV PATH="/usr/lib/postgresql/${POSTGRES_VERSION}/bin:${PATH}"
RUN pg_ctl -D /var/lib/postgresql/${POSTGRES_VERSION}/tests -l /var/log/postgresql/tests.log init

ENTRYPOINT ["/bin/bash"]