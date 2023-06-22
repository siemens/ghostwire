/*
Package mobydig implements deriving the DNS names of Docker services and
containers and then resolving and pinging them, based on the discovered network
topology, as well as Docker containers and Docker networks. The discovered DNS
names, resolved addresses, and address verification results are then streamed
via a “verdict” channel for consumption by, for instance, a client application.
*/
package mobydig
