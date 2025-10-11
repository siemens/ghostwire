/*
Package xdpnetdev provides setting up netdevs to work with XDP sockets,
especially loading the required dispatcher XDP program.

# Note

This is just a proof-of-concept and for testing, thus it does not support the
newer [libxdp] “version 2” XDP component program loading protocol.

[libxdp]: https://github.com/xdp-project/xdp-tools/tree/master/lib/libxdp
*/
package xdpnetdev
