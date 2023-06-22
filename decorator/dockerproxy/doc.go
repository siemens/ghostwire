/*
Package dockerproxy implements port forwarding detection based on Docker's
docker-proxy processes.

Using a [decorator.Decorator] is a bit of a stretch as decorators are intended
to decorate containers, but we can still reach the other stuff. And it is a
convenient modularization feature.
*/
package dockerproxy
