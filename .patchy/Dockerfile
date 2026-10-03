# The image patchy's agent runs in for this repository: patchy's agent base
# plus the pinned Go toolchain and this module's dependencies, offline. No
# application source is baked in. agent-image.yml builds it from main
# and the agent publisher pushes it as the tag .patchy/agent.yaml declares.
#
# Tags are immutable: when you change this file, or go.mod's dependencies,
# bump the toolchain-v<N> tag in .patchy/agent.yaml in the same commit. The
# agent has no network, so a module missing from the baked cache fails its
# build with "module lookup disabled by GOPROXY=off".
FROM golang:1.26.6@sha256:0d1d3a794be25f809dd2cb3160d8c73276c4056a9f8242a138e908ddeee7b6b6 AS toolchain

# The module cache, from go.mod and go.sum alone, in a stage of its own so
# neither file ends up in the image: an unchanged dependency set rebuilds to
# the same image.
FROM toolchain AS deps
ENV GOTOOLCHAIN=local GOMODCACHE=/opt/go/pkg/mod
WORKDIR /src
COPY go.mod go.sum* ./
RUN mkdir -p /opt/go/pkg/mod && go mod download

# patchy's agent base for the release the scaffold came from, pinned by
# digest: bash, git and the GNU tools, uid 65532, caches under /tmp.
FROM ghcr.io/devthenet-labs/patchy/agent-base:v0.12.15@sha256:1277da34f318d90c62f2461284c7eba666699f3fa91405ef743bbbf2f16a1cf2
COPY --from=toolchain /usr/local/go /usr/local/go
COPY --from=deps /opt/go/pkg/mod /opt/go/pkg/mod
# Offline from here on: GOPROXY=off resolves modules from the baked cache
# only, and -mod=mod lets go build and go test record a requirement the
# agent's change adds, provided the cache already holds that module.
ENV PATH=/usr/local/go/bin:${PATH} \
    GOTOOLCHAIN=local \
    GOMODCACHE=/opt/go/pkg/mod \
    GOPROXY=off \
    GOFLAGS=-mod=mod \
    CGO_ENABLED=0
LABEL org.opencontainers.image.source="https://github.com/devthenet-labs/app-template" \
      org.opencontainers.image.title="app-template agent toolchain"
WORKDIR /workspace
USER 65532:65532
