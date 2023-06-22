# Notes

Miscellaneous technical notes.

## nerdctl

Gostwire supports of the following [nerdctl](https://github.com/containerd)
features on top of `containerd`:

- 🏷️ `nerdctl/name` container label to specify a container's (functional) name,
  as opposed to a container instance-unique ID.

- nerdctl-managed CNI networks, as (dynamically) configured in
  `/etc/cni/net.d/nerdctl-*.conflist`. For more background details, please see
  the dedicated [nerdctl](nerdctl) section in this documentation.

## Turtle Finders

The plugin group type is `ghostwire.turtlefinder.detect.Detector`. Please see
[@thediveo/go-plugger](https://github.com/thediveo/go-plugger) for details on
the plugin mechanism used in Ghostwire.
