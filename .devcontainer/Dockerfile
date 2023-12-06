ARG FROM_REPOSITORY_DEVTOOLS=public.ecr.aws/r9i9c8e1
ARG FROM_REPOSITORY_ORTOOLS=public.ecr.aws/r9i9c8e1

ARG FROM_IMAGE_DEVTOOLS=who-dances-what-devtools
ARG FROM_IMAGE_ORTOOLS=who-dances-what-ortools

ARG TAG=latest

FROM ${FROM_REPOSITORY_DEVTOOLS}/${FROM_IMAGE_DEVTOOLS}:${TAG} as devtools

FROM ${FROM_REPOSITORY_ORTOOLS}/${FROM_IMAGE_ORTOOLS}:${TAG} as ortools

FROM node:21.4.0 as node

FROM golang:1.21.5-bookworm

COPY --from=devtools \
    /usr/local/bin/ \
    /usr/local/bin/

COPY --from=ortools \
    /usr/local/include/ \
    /usr/local/include/

COPY --from=ortools \
    /usr/local/lib/ \
    /usr/local/lib/

COPY --from=ortools \
    /usr/local/bin/ \
    /usr/local/bin/

RUN apt-get update && apt-get install -y \
    build-essential \
    catch2 \
    check \
    cmake \
    curl \
    gdb \
    git \
    less \
    locales-all \
    man \
    pkg-config \
    sudo \
    vim \
    zsh \
    zplug \
    # Runtime deps for OR-Tools
    libeigen3-dev \
    libre2-dev \
    zlib1g-dev

COPY --from=node \
    /usr/local/bin/ \
    /usr/local/bin/

COPY --from=node \
    /usr/local/lib/ \
    /usr/local/lib/

COPY --from=node \
    /usr/local/include/ \
    /usr/local/include/

COPY --from=node \
    /usr/local/share/doc \
    /usr/local/share/doc

COPY --from=node \
    /usr/local/share/man \
    /usr/local/share/man

RUN useradd -m devcontainer -s /bin/zsh
RUN echo "devcontainer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/devcontainer

USER devcontainer
WORKDIR /home/devcontainer
ENV HOME /home/devcontainer
RUN echo "set debuginfod enabled on" >> ~/.gdbinit
ENV DEBUGINFOD_URLS="https://debuginfod.ubuntu.com"

ENV GOPATH="${HOME}/go"
ENV PATH="/go/bin:/usr/local/go/bin:${GOPATH}/bin:${PATH}"

SHELL ["/bin/zsh", "-c"]
