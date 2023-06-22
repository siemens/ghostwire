#!/bin/bash

# Download aux tools (nerdctl, CNI plugins) required to run some of the Gostwire
# tests.
#
# On purpose, we're not going to install from often totally outdated or lacking
# distro repositories, but from the Github projects directly.

CNI_PLUGINS_VERSION=v1.3.0
NERDCTL_VERSION=v1.4.0

BINDIR=${BINDIR:-/usr/local/bin}
CNIBINDIR=${CNIBINDIR:-/opt/cni/bin}

ARCH=$(dpkg --print-architecture)
echo "downloading tools for architecture ${ARCH}"

WGET_PREX="/tmp/nerdctl-and-friends"
mkdir -p ${WGET_PREX}
wget -P ${WGET_PREX} --backups=1 https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz
wget -P ${WGET_PREX} --backups=1 https://github.com/containerd/nerdctl/releases/download/${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION#v}-linux-${ARCH}.tar.gz

echo "installing CNI plugins into ${CNIBINDIR}"
mkdir -p ${CNIBINDIR}
sudo tar -x -v -z -f ${WGET_PREX}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz -C ${CNIBINDIR}
echo "installing nerdctl into ${BINDIR}"
sudo tar -x -v -z -f ${WGET_PREX}/nerdctl-${NERDCTL_VERSION#v}-linux-${ARCH}.tar.gz -C ${BINDIR}
