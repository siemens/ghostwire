/*
Package iecore implements a metadata plugin that returns information about the
Industrial Edge (“core”) runtime container, if present.

The (JSON) metadata is structured (inside “metadata”) as follows:

  - “industrial-edge”: object
  - “semversion"” string, semantic version

The IE core/runtime version is read from the container's
/etc/os-release-container file, and its VERSION_ID variable in particular.
Please note that this plugin silently fixes the broken “-e VERSION_ID” variable
name present in some versions of the runtime container.
*/
package iecore
