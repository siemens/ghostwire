# ⚠️ Requires Docker Buildx in order to correctly produce the final image with
# the extended file attributes for capabilities still intact. The Buildx plugin
# is automatically installed on recent .deb and .rpm based distributions.

# TAGS can be used to pass additional Go build tags to the service binary build
# step.

# Nota bene: not all existing value combinations with respect to Alpine and Node
# versions might work, so you'll need to check Docker hub for the available base
# image versions.
ARG ALPINE_VERSION=3.19
ARG ALPINE_PATCH=1
ARG GO_VERSION=1.21.7
ARG NODE_VERSION=21

# Go build settings
ARG LDFLAGS="-s -w -extldflags=-static"
ARG TAGS="osusergo,netgo,sqlite_omit_load_extension"

# 0th stage: https://github.com/tonistiigi/xx/blob/master/README.md
FROM --platform=${BUILDPLATFORM} tonistiigi/xx AS cc-helpers

# -- 1st stage ---------------------------------------------------------------
#
# Build the gostwire binary, now this requires cgo and we thus need a C (cross-)
# compiler ... and then we also need header files. Caching to the rescue at
# least for local builds, so we start with the C-related stuff, which is kind of
# base builder image stuff anyway.
#
# Please note that the gostwire binary expects the web UI files to be later made
# available in the final image. We will later reuse this stage to build the
# final gostwire binary with the web UI files embedded.
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ARG LDFLAGS
ARG TAGS

ARG BUILDPLATFORM
ARG TARGETPLATFORM

ARG WSDISKPATHS
ARG MOD1=./
ARG MOD2=./
ARG MOD3=./
ARG MOD4=./
ARG MOD5=./
ARG MOD6=./
ARG MOD7=./
ARG MOD8=./
ARG MOD9=./

# General warm-up
RUN apk add --no-cache git clang lld libcap-utils
COPY --from=cc-helpers / /
RUN xx-apk add --no-cache gcc musl-dev
# https://github.com/tonistiigi/xx/blob/master/README.md#go--cgo
ENV CGO_ENABLED=1
# https://github.com/mattn/go-sqlite3/issues/1164
ENV CGO_CFLAGS=-D_LARGEFILE64_SOURCE
# As of Go 1.20 the Go distribution won't contain a pre-build std library
# anymore, so do this now and cache it.
RUN xx-go build std

WORKDIR /ws
# Copy the additionally used modules into the soon-to-be workspace.
COPY --from=bctx1 . ${MOD1}
COPY --from=bctx2 . ${MOD2}
COPY --from=bctx3 . ${MOD3}
COPY --from=bctx4 . ${MOD4}
COPY --from=bctx5 . ${MOD5}
COPY --from=bctx6 . ${MOD6}
COPY --from=bctx7 . ${MOD7}
COPY --from=bctx8 . ${MOD8}
COPY --from=bctx9 . ${MOD9}

# Make sure we have the main module containing a main package to be build...
COPY go.mod go.sum ./ghostwire/

# Establish the Go workspace
RUN go work init ${WSDISKPATHS}

WORKDIR /ws/ghostwire
# We now try to cache only the dependencies in a separate layer, so we can speed
# up things in case the dependencies do not change. This then reduces the amount
# of fetching and compiling required when compiling the final binary later.
RUN go mod download
# And now, finally, we build the gostwire service itself.
COPY api/ ./api/
COPY cmd/ ./cmd/
COPY decorator/ ./decorator/
COPY internal/ ./internal/
COPY metadata/ ./metadata/
COPY mobydig/ ./mobydig/
COPY network/ ./network/
COPY *.go ./

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    xx-go build -v \
        -tags="${TAGS}" \
        -ldflags="${LDFLAGS}" \
        -o /gostwire ./cmd/gostwire && \
    xx-verify --static /gostwire

# -- 2nd stage ---------------------------------------------------------------
#
# Build Ghostwire's frontend UI browser application.
FROM --platform=${BUILDPLATFORM} node:${NODE_VERSION}-alpine AS webui
WORKDIR /webui
ENV PATH /webui/node_modules/.bin:$PATH
RUN yarn set version berry
RUN yarn config set nodeLinker node-modules
# stage 2.1 :) install a full-blown node_modules dependency hell; we need to
# tell yarn to be more patient. This gives us a nice cached layer.
COPY --from=webappsrc \
    package.json \
    yarn.lock \
    .eslintrc.cjs \
    tsconfig.json \
    tsconfig.node.json \
    vite.config.ts \
    index.html \
    ghostwire.svg \
        ./

RUN yarn workspaces focus --production
RUN yarn eslint --init

# stage 2.2 only now copy in the web ui completely and build the optimized
# static artefacts.
COPY --from=webappsrc public/ ./public/
COPY --from=webappsrc src/ ./src/
COPY --from=webappsrc .env ./
ARG REACT_APP_GIT_VERSION
RUN sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=${REACT_APP_GIT_VERSION}/" .env
RUN yarn build

# -- 3. stage ---------------------------------------------------------------
#
# Complete the gostwire binary by embedding the web UI files this time.
FROM --platform=${BUILDPLATFORM} builder AS embedder

ARG LDFLAGS
ARG TAGS

ARG BUILDPLATFORM
ARG TARGETPLATFORM

WORKDIR /ws/ghostwire

# Now we need the webui package that simply consists of a go:embed directive and
# the Webui fs.FS export.
COPY --from=webappsrc webui.go ./webui/
COPY --from=webui /webui/build/ ./webui/build/
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    xx-go build -v \
        -tags="${TAGS},webui" \
        -ldflags="${LDFLAGS}" \
        -o /gostwire ./cmd/gostwire && \
    xx-verify --static /gostwire
RUN setcap "cap_sys_admin,cap_sys_chroot,cap_sys_ptrace,cap_dac_read_search,cap_dac_override,cap_net_raw,cap_net_admin+ep" /gostwire

# -- 4. stage ---------------------------------------------------------------
#
# In the final stage we now "only" need to install the gostwire service binary
# and its web ui application into the final image.
FROM alpine:${ALPINE_VERSION}.${ALPINE_PATCH} as final
LABEL maintainer="Harald Albrecht <harald.albrecht@siemens.com>"
ARG ALPINE_VERSION
ARG GOLANG_VERSION
COPY --from=embedder /gostwire /gostwire
#COPY --from=webui /webui/build /webui/build
# Now bugger off...
USER 65534
# The ENTRYPOINT specifies the program to execute, as well as the "fixed"
# command line arguments.
ENTRYPOINT ["/gostwire"]
# CMD in combination with ENTRYPOINT now specifies *additional* default command
# line arguments, which will be added to the ENTRYPOINT parameters. These
# default arguments can be easily overriden when using "docker run <container>
# additional-args".
CMD ["--http", "[::]:5000"]
