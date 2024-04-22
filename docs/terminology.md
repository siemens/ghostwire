<!-- markdownlint-disable MD001 -->
# Terminology

#### Containerizer

A **containerizer** discovers containers in a host system. The
[lxkns](https://github.com/thediveo/lxkns) module then relates these discovered
containers to discovered namespaces and processes. lxkns itself comes with a
built-in Containerizer that handles a static and pre-configured set of container
engine watchers, namely: a Docker host daemon and a containerd host daemon.

Gostwire supplies its own Containerizer that features dynamic discovery of
container engines, whereever they are: in the host, in containers, et cetera.
The engine detection bases on daemon process names.

#### Decorator

1. A ([lxkns](https://github.com/thediveo/lxkns)) plugin that post-processes the
   containers discovered by a Containerizer. For instance, to detect Docker
   compose projects and then to group containers into their respective composer
   project groups. The Decorator plugin group name is
   `"lxkns/plugingroup/decorator"`, defined as constant
   [`decorator.PluginGroup`](https://github.com/TheDiveO/lxkns/blob/master/decorator/decorator.go#L21).

2. A Gostwire plugin that post-processes the discovered network namespaces. For
   instance, to gather the names of Docker networks, et cetera.

#### Detector

A [turtlefinder](https://github.com/siemens/turtlefinder) plugin that detects if
a given process is a particular container engine and then contacts its API for
discovering the containers this particular engine manages. The Detector plugin
group name is `"turtlefinders"`.
