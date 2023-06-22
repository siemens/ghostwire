#!/bin/bash
set -e
UNDERLAY=$(ls -l /sys/class/net/ | grep -v virtual | awk '{print $9}' | grep -E '^(en|eth)')
VXLANERPID=$(docker inspect -f '{{.State.Pid}}' mock_vxlaner_1)

# creates VXLAN overlay and its underlay in the initial network namespace, if it
# doesn't exist yet.
if ! ip -o link show vxtestlan123 >/dev/null 2>&1; then
    sudo ip link add vxtestlan123 type vxlan id 12346 dstport 4789 dev $UNDERLAY
    echo "created vxtestlan123 in host"
else
    echo "vxtestlan123 already exists"
fi

# creates VXLAN overlay in a different network namespace (of the
# "mock_vxlaner_1" container), if it doesn't exist yet.
sudo mkdir -p /var/run/netns
sudo ln -sf /proc/$VXLANERPID/ns/net /var/run/netns/vxlaner
if ! sudo ip -n vxlaner link show vxtestlan >/dev/null 2>&1; then
    # creates overlay interface in host first (as otherwise we cannot reference
    # the underlay network interface), then move the overlay interface into its
    # container.
    sudo ip link add vxtestlan type vxlan id 12345 dstport 4789 dev $UNDERLAY
    sudo ip link set vxtestlan netns vxlaner
    echo "created VXLAN overlay 12345 access in container mock_vxlaner_1"
else
    echo "container mock_vxlaner_1 has already access to VXLAN overlay 12345"
fi
sudo rm -f /var/run/netns/vxlaner
