# Build our Gostwire Docker image as a distroless-like image: a static Gostwire
# binary with embedded web UI, plus only a minimal root filesystem unpacked
# from a fixed Debian stable snapshot, without any shell or package manager.
#
# ⚠️ Requires Docker Buildx in order to correctly produce the final image with
# the extended file attributes for capabilities still intact. The Buildx plugin
# is automatically installed on recent .deb and .rpm based distributions.

# TAGS can be used to pass additional Go build tags to the service binary build
# step.

# Debian packages come from snapshot.debian.org as of this point in time, so
# rebuilds get the same package versions; use "make bump-debian" to pick up
# (security) updates. An empty DEBIAN_SNAPSHOT uses the live Debian mirrors
# instead, as a fallback when snapshot.debian.org is down.
ARG DEBIAN_SNAPSHOT=20261105T143213Z

ARG NODE_VERSION=24
# keep in sync with "packageManager" in webui/package.json
ARG YARN_VERSION=4.18.1

# Go build settings
ARG LDFLAGS="-s -w -extldflags=-static"
ARG TAGS="osusergo,netgo,sqlite_omit_load_extension"

# VERSION is the (git describe/tag) version to stamp the binary and web UI with.
ARG VERSION=""

# https://github.com/tonistiigi/xx/blob/master/README.md
FROM --platform=${BUILDPLATFORM} tonistiigi/xx:1.10.0@sha256:40fe21f44bb8cddf1eb1226db190d0a984118f62ee02a91b77d99448d28264a6 AS cc-helpers

# -- Debian base -------------------------------------------------------------
#
# Debian stable with the Go toolchain, and its package sources pointing to our
# fixed snapshot.
FROM --platform=${BUILDPLATFORM} golang:1.26-trixie@sha256:9ce8fcba4b6fa1ae0fc1c00d255c073cfb9f9783237bb6315512fed4051d83ce AS debian

ARG DEBIAN_SNAPSHOT

SHELL ["/bin/bash", "-o", "pipefail", "-ec"]
COPY --from=cc-helpers / /
# Plain http is fine, as apt checks the signed Release files; these will have
# expired for older snapshots, though. As snapshot.debian.org is flaky, give up
# on stalled connections early and retry, and fail on index fetch errors, never
# silently caching incomplete lists.
RUN if [ -n "${DEBIAN_SNAPSHOT}" ]; then \
        sed -i -E "s#^URIs: http://deb.debian.org/(debian(-security)?)\$#URIs: http://snapshot.debian.org/archive/\1/${DEBIAN_SNAPSHOT}#" \
            /etc/apt/sources.list.d/debian.sources && \
        echo 'Acquire::Check-Valid-Until "false";' > /etc/apt/apt.conf.d/99snapshot; \
    fi && \
    printf 'Acquire::http::Timeout "15";\nAcquire::Retries "5";\nAPT::Update::Error-Mode "any";\n' >> /etc/apt/apt.conf.d/99snapshot

# Empty defaults for the additional Go workspace module build contexts, unless
# overridden using --build-context bctxN=...
FROM scratch AS bctx1
FROM scratch AS bctx2
FROM scratch AS bctx3
FROM scratch AS bctx4
FROM scratch AS bctx5
FROM scratch AS bctx6
FROM scratch AS bctx7
FROM scratch AS bctx8
FROM scratch AS bctx9

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
FROM debian AS builder

ARG LDFLAGS
ARG TAGS

ARG TARGETPLATFORM

ARG WSDISKPATHS=./ghostwire
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
RUN apt-get update && \
    apt-get install -y --no-install-recommends clang lld libcap2-bin && \
    xx-apt-get install -y --no-install-recommends gcc libc6-dev
# https://github.com/tonistiigi/xx/blob/master/README.md#go--cgo
# cgo is required by lxkns' mntnssandbox (C constructor does setns before the Go runtime starts).
ENV CGO_ENABLED=1
# As of Go 1.20 the Go distribution won't contain a pre-build std library
# anymore, so do this now and cache it.
RUN --mount=type=cache,target=/root/.cache/go-build \
    xx-go build std

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

# Make sure we have the main module containing a main package to be built...
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
COPY innetns/ ./innetns/
COPY internal/ ./internal/
COPY meta/ ./meta/
COPY mobydig/ ./mobydig/
COPY netdev/ ./netdev/
COPY network/ ./network/
COPY passedthrough/ ./passedthrough/
COPY xsk/ ./xsk/
COPY *.go ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    xx-go build -v \
    -trimpath \
    -tags="${TAGS}" \
    -ldflags="${LDFLAGS}" \
    -o /gostwire ./cmd/gostwire && \
    xx-verify --static /gostwire

# -- 2nd stage ---------------------------------------------------------------
#
# Build Ghostwire's frontend UI browser application.
FROM --platform=${BUILDPLATFORM} node:${NODE_VERSION}-alpine AS webui
WORKDIR /webui
ENV PATH=/webui/node_modules/.bin:$PATH
ARG YARN_VERSION
RUN yarn set version ${YARN_VERSION} && \
    yarn config set nodeLinker node-modules
# stage 2.1 :) install a full-blown node_modules dependency hell; we need to
# tell yarn to be more patient. This gives us a nice cached layer.
COPY --from=webappsrc package.json yarn.lock ./

#RUN yarn workspaces focus --production
RUN yarn

# stage 2.2 only now copy in the web ui completely and build the optimized
# static artefacts.
COPY --from=webappsrc \
    eslint.config.js \
    tsconfig.json \
    tsconfig.app.json \
    tsconfig.node.json \
    vite.config.ts \
    index.html \
    ghostwire.svg \
    ./
COPY --from=webappsrc public/ ./public/
COPY --from=webappsrc src/ ./src/
COPY --from=webappsrc .env ./
ARG VERSION
RUN sed -i "s/^VITE_REACT_APP_GIT_VERSION=.*/VITE_REACT_APP_GIT_VERSION=${VERSION#v}/" .env
RUN yarn build

# -- 3. stage ---------------------------------------------------------------
#
# Complete the gostwire binary by embedding the web UI files this time.
FROM builder AS embedder

ARG LDFLAGS
ARG TAGS
ARG VERSION

ARG TARGETPLATFORM

WORKDIR /ws/ghostwire

# Now we need the webui package that simply consists of a go:embed directive and
# the Webui fs.FS export.
COPY --from=webappsrc webui.go ./webui/
COPY --from=webui /webui/build/ ./webui/build/
RUN --mount=type=cache,target=/root/.cache/go-build \
    xx-go build -v \
    -trimpath \
    -tags="${TAGS},webui" \
    -ldflags="${LDFLAGS} -X github.com/siemens/ghostwire/v2.SemVersion=${VERSION#v}" \
    -o /gostwire ./cmd/gostwire && \
    xx-verify --static /gostwire
RUN setcap "cap_sys_admin,cap_sys_chroot,cap_sys_ptrace,cap_dac_read_search,cap_dac_override,cap_net_raw,cap_net_admin+ep" /gostwire

# -- 4. stage ---------------------------------------------------------------
#
# Assemble the root filesystem of the final image from Debian stable packages,
# for the target architecture but without running anything under emulation.
# Instead of installing the packages, we only unpack them: this skips their
# maintainer scripts. As gostwire is a static binary, no shared libraries are
# needed.
FROM debian AS rootfs

ARG TARGETPLATFORM

RUN dpkg --add-architecture "$(xx-info debian-arch)" && \
    apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates

WORKDIR /debs
RUN apt-get download "base-files:$(xx-info debian-arch)" ca-certificates tzdata

# Like distroless, add package records for vulnerability scanners, and keep
# only the copyright files from the documentation.
RUN mkdir -p /rootfs && \
    dpkg-deb -x base-files_*.deb /rootfs && \
    mkdir -p /rootfs/var/lib/dpkg/status.d && \
    for deb in *.deb; do \
        pkg="$(dpkg-deb -f "${deb}" Package)"; \
        dpkg-deb -f "${deb}" > "/rootfs/var/lib/dpkg/status.d/${pkg}"; \
        dpkg-deb --ctrl-tarfile "${deb}" | tar -xO ./md5sums > "/rootfs/var/lib/dpkg/status.d/${pkg}.md5sums"; \
        case "${pkg}" in \
        base-files) ;; \
        ca-certificates) \
            dpkg-deb -x "${deb}" /tmp/pkg; \
            install -D -m 0644 /tmp/pkg/usr/share/doc/${pkg}/copyright /rootfs/usr/share/doc/${pkg}/copyright; \
            install -D -m 0644 /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/ca-certificates.crt; \
            rm -rf /tmp/pkg;; \
        *) dpkg-deb -x "${deb}" /rootfs;; \
        esac; \
    done && \
    find /rootfs/usr/share/doc -type f ! -name copyright -delete && \
    rm -rf /rootfs/usr/share/man && \
    printf 'root:x:0:0:root:/root:/usr/sbin/nologin\nnobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin\n' > /rootfs/etc/passwd && \
    printf 'root:x:0:\nnogroup:x:65534:\n' > /rootfs/etc/group && \
    test -L /rootfs/bin && test -L /rootfs/lib && test -L /rootfs/sbin

# -- final stage -------------------------------------------------------------
#
# In the final stage we now "only" need to install the gostwire service binary
# with its embedded web ui application on top of the minimal root filesystem.
FROM scratch AS final
COPY --from=rootfs /rootfs/ /
COPY --from=embedder /gostwire /gostwire
EXPOSE 5000
# Now bugger off...
USER 65534
# The ENTRYPOINT specifies the program to execute, as well as the "fixed"
# command line arguments.
ENTRYPOINT ["/gostwire"]
# CMD in combination with ENTRYPOINT now specifies *additional* default command
# line arguments, which will be added to the ENTRYPOINT parameters. These
# default arguments can be easily overridden when using "docker run <container>
# additional-args".
CMD ["--http", "[::]:5000"]
# Labels last, so that a changing VERSION doesn't invalidate the layers above.
ARG VERSION
LABEL org.opencontainers.image.title="gostwire" \
      org.opencontainers.image.description="Container virtual networks discovery service" \
      org.opencontainers.image.source="https://github.com/siemens/ghostwire" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="Harald Albrecht <harald.albrecht@siemens.com>" \
      org.opencontainers.image.version="${VERSION}"
