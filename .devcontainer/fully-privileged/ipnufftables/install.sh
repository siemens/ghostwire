#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

echo "Activating feature 'ipnufftables'..."

if [ "$(id -u)" -ne 0 ]; then
    err 'Script must be run as root. Use sudo, su, or add "USER root" to your Dockerfile before running this script.'
    exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
    echo "This feature currently supports only Debian/Ubuntu-based images."
    exit 1
fi

require() {
    local description="$1"
    shift

    if ! "$@" >/dev/null 2>&1; then
        echo "Requirement failed: ${description}"
        echo "Command: $*"
        exit 1
    fi
}

apt-get update
apt-get install -y --no-install-recommends \
    iptables \
    nftables \
    ca-certificates

require "/usr/sbin/iptables-nft exists and is executable" test -x /usr/sbin/iptables-nft
require "/usr/sbin/ip6tables-nft exists and is executable" test -x /usr/sbin/ip6tables-nft
require "update-alternatives entry for iptables exists" update-alternatives --query iptables
require "update-alternatives entry for ip6tables exists" update-alternatives --query ip6tables

update-alternatives --set iptables /usr/sbin/iptables-nft
update-alternatives --set ip6tables /usr/sbin/ip6tables-nft

if [ -x /usr/sbin/arptables-nft ]; then
    require "update-alternatives entry for arptables exists" update-alternatives --query arptables
    update-alternatives --set arptables /usr/sbin/arptables-nft
fi

if [ -x /usr/sbin/ebtables-nft ]; then
    require "update-alternatives entry for ebtables exists" update-alternatives --query ebtables
    update-alternatives --set ebtables /usr/sbin/ebtables-nft
fi

echo "Verifying backend..."
iptables --version || true
ip6tables --version || true
nft --version || true

rm -rf /var/lib/apt/lists/*

echo "Done."
