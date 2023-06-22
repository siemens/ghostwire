/*
Package turtlefinder provides a Containerizer that auto-detects container
engines and their containers. The turtlefinder containerizer can be used by
concurrent discoveries.

All that is necessary:

	containerizer := turtlefinders.New()

Boringly simple, right?

Basically, upon a container query the turtlefinder containerizer first looks for
any newly seen container engines, based on container engine process names. The
engine discovery can be extended by pluging in new engine detectors (and
adaptors). Additionally, the turtlefinder determines the hierarchy of container
engines, such as when a container engine is hosted inside a container managed by
a (parent) container engine. This hierarchy later gets propagated to the
individual containers in form of a so-called “prefix”, attached in form of a
special container label.

The turtlefinder then spins up background watchers as required that synchronize
with the workload states. Old engine watchers get retired as their engine
processes die. This workload state information is then returned as the list of
discovered containers.

The decoration of the discovered containers is then done as usual by the
(extensible) lxkns decorator mechanism as part of the overall discovery.
*/
package turtlefinder
