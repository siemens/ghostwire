/*
Package nerdctlnet implements a Gostwire decorator that discovers
nerdctl-managed CNI networks and then decorates their corresponding network
interfaces. Supported types of nerdctl/CNI networks are: “bridge”. Support for
“macvlan” is currently blocked upstream in nerdctl as well the CNI plugins.

In case of “bridge” networks this decorator assigns nerdctl/CNI network names as
alias names to the corresponding Linux-kernel bridges and also as a
Gostwire-specific label.

This decorator also copies any network labels it finds into the corresponding
network.Interface instances in a Gostwire discovery information model.
*/
package nerdctlnet
