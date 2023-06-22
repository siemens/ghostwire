#!/bin/bash
set -e

# Download aux tools (nerdctl, CNI plugins) required to run some of the Gostwire
# tests.
#
# On purpose, we're not going to install from often totally outdated or "sparse"
# distro repositories, but from the Github projects directly.

CNI_PLUGINS_VERSION=v1.3.0
NERDCTL_VERSION=v1.4.0

WGET_PREX="/tmp/nerdctl-and-friends"

BINDIR=${BINDIR:-/usr/local/bin}
CNIBINDIR=${CNIBINDIR:-/usr/lib/cni}

ARCH=$(dpkg --print-architecture)
echo "downloading tools for architecture ${ARCH}"

mkdir -p ${WGET_PREX}
trap 'rm -rf -- "${WGET_PREX}"' EXIT

wget -q -P ${WGET_PREX} --backups=1 https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz
wget -q -P ${WGET_PREX} --backups=1 https://github.com/containerd/nerdctl/releases/download/${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION#v}-linux-${ARCH}.tar.gz

echo "installing CNI plugins into ${CNIBINDIR}"
mkdir -p ${CNIBINDIR}
sudo tar -x -v -z -f ${WGET_PREX}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz -C ${CNIBINDIR}
echo "installing nerdctl into ${BINDIR}"
sudo tar -x -v -z -f ${WGET_PREX}/nerdctl-${NERDCTL_VERSION#v}-linux-${ARCH}.tar.gz -C ${BINDIR}
