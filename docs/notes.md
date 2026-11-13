# Notes

Miscellaneous technical notes.

## KinD

[Kubernetes in Docker (KinD)](https://kind.sigs.k8s.io/) runs Kubernetes
clusters with each node being a Docker container that in turn runs its own
`containerd` container engine. **Ghostwire** detects and understands KinD: it
automatically discovers the `containerd` engines inside the KinD node containers
and correctly nests the pods' containers beneath their respective node
containers.

## nerdctl

**Ghostwire** supports of the following [nerdctl](https://github.com/containerd)
features on top of `containerd`:

- 🏷️ `nerdctl/name` container label to specify a container's (functional) name,
  as opposed to a container instance-unique ID.

- nerdctl-managed CNI networks, as (dynamically) configured in
  `/etc/cni/net.d/nerdctl-*.conflist`. For more background details, please see
  the dedicated [nerdctl](nerdctl) section in this documentation.

## Turtle Finders

The plugin group type is `ghostwire.turtlefinder.detect.Detector`. Please see
[@thediveo/go-plugger](https://github.com/thediveo/go-plugger) for details on
the plugin mechanism used in **Ghostwire**.
