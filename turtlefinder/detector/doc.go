/*
Package detector defines the plugin interface between the TurtleFinder and its
container engine detector plugins.

The sub-package “all” pulls in all supported engine detector plugins, that are
supported out-of-the-box. The individual engine-specific detector plugins are
then implemented in the other sub-packages: for instance, the “containerd” and
“moby” sub-packages.
*/
package detector
