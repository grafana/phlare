# syntax=docker/dockerfile:1.4

# NOTE: This Dockerfile can only be built using BuildKit. BuildKit is used by
# default when running `docker buildx build` or when DOCKER_BUILDKIT=1 is set
# in environment variables.

FROM --platform=$BUILDPLATFORM grafana/agent-build-image:0.26.0 as build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG RELEASE_BUILD=1
ARG VERSION

COPY . /src/
WORKDIR /src/

# Build the UI before building the agent, which will then bake the final UI
# into the binary.
#RUN --mount=type=cache,target=/src/agent/web/ui/node_modules,sharing=locked \
#   make generate-ui

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    RELEASE_BUILD=${RELEASE_BUILD} VERSION=${VERSION}  \
    CGO_ENABLED=0  \
    go build -ldflags "-s -w"  -tags "netgo " -o tinyagent ./ebpf/cmd/tinyagent

FROM ubuntu:lunar

LABEL org.opencontainers.image.source="https://github.com/grafana/agent"

# Install dependencies needed at runtime.
RUN <<EOF
  apt-get update
  apt-get install -qy libsystemd-dev tzdata ca-certificates

  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
EOF

COPY --from=build /src/tinyagent /bin/tinyagent


ENTRYPOINT ["/bin/tinyagent"]
