/*
Package initialcgroup provides the “--initialcgroup” CLI flag that switches the
Gostwire service into the initial cgroup namespace, working around Docker
compose deployments that cannot set “cgroupns: host”.

When the flag has been specified, the service switches into the initial cgroup
namespace and then re-executes itself, before the command runs. The
re-executed service gets the additional hidden “--cgroupswitched” flag, so it
knows to not switch again.
*/
package initialcgroup
