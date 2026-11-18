#!/bin/bash
# (c) Siemens AG 2026
#
# SPDX-License-Identifier: MIT

# Smoke tests a container image built from this example downstream module: the
# image must keep the upstream image configuration, run the composed service,
# and serve its own endpoints, the web UI, and the Gostwire API with the
# additional decorator applied.
#
# Usage: smoketest.sh IMAGE
set -euo pipefail

image="${1:?usage: smoketest.sh IMAGE}"
name="gostwire-downstream-smoketest"
port="${SMOKETEST_PORT:-5097}"
repo="$(realpath "$(dirname "$0")/../..")"

fail() { echo "❌ $*" >&2; exit 1; }
ok() { echo "✅ $*"; }

expect_eq() {
    [[ "$2" == "$3" ]] || fail "$1: expected '$3', got '$2'"
    ok "$1"
}

expect_eq "entrypoint" "$(docker inspect -f '{{json .Config.Entrypoint}}' "${image}")" '["/gostwire"]'
expect_eq "cmd" "$(docker inspect -f '{{json .Config.Cmd}}' "${image}")" '["--http","[::]:5000"]'
expect_eq "user" "$(docker inspect -f '{{.Config.User}}' "${image}")" "65534"
expect_eq "title label" \
    "$(docker inspect -f '{{index .Config.Labels "org.opencontainers.image.title"}}' "${image}")" "gostwire"

# Restoring file capabilities requires root, so extract the service binary
# with sudo, unless already running as root.
sudo=""
[[ "$(id -u)" == 0 ]] || sudo="sudo -n"
tmp="$(mktemp -d)"
cid="$(docker create "${image}")"
trap 'docker rm -f "${name}" "${cid}" >/dev/null 2>&1 || true; ${sudo} rm -rf -- "${tmp}"' EXIT
docker export "${cid}" | ${sudo} tar -x --xattrs --xattrs-include='security.*' -C "${tmp}" gostwire
expect_eq "capabilities" "$(${sudo} getcap "${tmp}/gostwire" | awk '{print $2}')" \
    "cap_dac_override,cap_dac_read_search,cap_net_admin,cap_net_raw,cap_sys_chroot,cap_sys_ptrace,cap_sys_admin=ep"

# The service binary's file capabilities must be in the bounding set.
runopts=(
    --read-only --user 65534 --pid host
    --cap-drop ALL
    --cap-add SYS_ADMIN --cap-add SYS_CHROOT --cap-add SYS_PTRACE
    --cap-add DAC_READ_SEARCH --cap-add DAC_OVERRIDE
    --cap-add NET_RAW --cap-add NET_ADMIN
    --security-opt apparmor=unconfined
    --security-opt seccomp="${repo}/deployments/gostwire/seccomp/gostwire.json"
)

version="$(docker run --rm "${runopts[@]}" "${image}" --version)"
[[ "${version}" == "gostwire-example version "* ]] || fail "unexpected version: ${version}"
ok "version: ${version}"

docker run --detach --name "${name}" "${runopts[@]}" \
    --publish "127.0.0.1:${port}:5000" \
    "${image}" >/dev/null
base="http://127.0.0.1:${port}"
for _ in $(seq 60); do
    curl -sf "${base}/api/v1/hello" >/dev/null && break
    sleep 0.5
done

hello="$(curl -sf "${base}/api/v1/hello")" || { docker logs "${name}"; fail "no /api/v1/hello"; }
[[ "${hello}" == *'"greeting":"hello"'* ]] || fail "unexpected /api/v1/hello: ${hello}"
ok "additional endpoint: ${hello}"

curl -sf "${base}/json" | grep -q '"example.org/hello":"world"' || fail "additional decorator not applied"
ok "additional decorator"

curl -sf -D - -o /dev/null "${base}/json" | grep -qi '^X-Example-Override: hello' || fail "upstream endpoint not overridden"
ok "overridden upstream endpoint"

curl -sf "${base}/mobyshark" | grep -q '"containers"' || fail "no /mobyshark"
ok "Gostwire API"

curl -sf "${base}/w" | grep -q 'window.dynvars = {"enableMonolith"' || fail "no default web UI"
ok "default web UI"
