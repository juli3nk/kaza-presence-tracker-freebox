ARG GO_VERSION="1.22"
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}

ARG GROUP_ID_DOCKER="989"

RUN mkdir -p /etc/apt/keyrings \
  && curl -sfSL https://download.docker.com/linux/debian/gpg | gpg --dearmor --yes --output /etc/apt/keyrings/docker.gpg \
  && chmod a+r /etc/apt/keyrings/docker.gpg \
  && echo \
    "deb [arch=\"$(dpkg --print-architecture)\" signed-by=\"/etc/apt/keyrings/docker.gpg\"] https://download.docker.com/linux/debian \
    $(awk -F'=' '/VERSION_CODENAME/ { print $2 }' /etc/os-release) stable" > /etc/apt/sources.list.d/docker.list \
  && apt update \
  && apt install --no-install-recommends --yes \
    docker-ce-cli \
    docker-compose-plugin \
  && groupadd -g "$GROUP_ID_DOCKER" -U vscode docker
