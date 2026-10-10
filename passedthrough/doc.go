/*
Package passedthrough provides detection of “original ifnames” as well as the
related Docker networks for host network interfaces roaming around other network
namespaces due to passthrough-like container networking drivers.

This detection is based on cooperative container networking drivers attaching
so-called “alternative names” (not “aliases”!) that start with a special prefix
and give passthrough details.
  - one alias describing the original network interface name before (temporary)
    renaming.
  - another alias describing the associated Docker network ID.

Please note that Linux allows attaching multiple [alternative names] to the same
network interface and maintaining them individually. In contrast, the older
Linux mechanism of alias names allow only a single alias per network interface.

Siemens Industrial Edge devices use a cooperating custom Docker passthrough
driver named “iedge”, as documented in [Network Interface Isolation].

[Network Interface Isolation]: https://docs.industrial-operations-x.siemens.cloud/r/en-us/v26.01/industrial-edge-platform-operation-develop-an-application/developer-guide/industrial-edge-platform/resource-manager/network-interface-isolation
[alternative names]: https://lwn.net/Articles/794289/
*/
package passedthrough
