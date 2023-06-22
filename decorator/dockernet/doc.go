/*
Package dockernet implements a Gostwire decorator that discovers Docker-managed
networks and then decorates their corresponding Linux-kernel network interfaces.
Supported types of Docker networks are “bridge” and “macvlan”.

In case of “bridge” networks this decorator assigns Docker network names as
alias names to the corresponding Linux-kernel bridges and also as a
Gostwire-specific label.

For “MACVLAN” networks this decorator assigns the Docker network names as alias
names to the “parent” network interface (or “master” in Linux parlance).

This decorator also copies any network labels it finds into the corresponding
network.Interface instances in a Gostwire discovery information model.
*/
package dockernet
