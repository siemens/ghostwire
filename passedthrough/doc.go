/*
Package passedthrough provides detection of “original ifnames” for host network
interfaces roaming around other network namespaces due to passthrough-like
container networking drivers.

This detection is based on cooperative container networking drivers attaching a
so-called “alternative name” that starts with a special prefix and then contains
the original network interface name before (temporary) renaming.

Please note that Linux allows attaching multiple alias names to the same network
interface and maintaining them individually.

[alternative name]: https://lwn.net/Articles/794289/
*/
package passedthrough
