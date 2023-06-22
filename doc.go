/*
Package gostwire provides discovery of topology and configuration of virtual
networks in Linux hosts.

# Discovery

The Discover function expects a context.Context as well as a so-called
“containerizer” implementing the containerizer.Containerizer interface.
Gostwire's turtlefinder package offers a suitable implementation, accessible as
turtlefinder.New(). This implementation offers automatic detection of container
engines without any need for API endpoint configuration. The containerizer needs
to be allocated only once and is safe for use in concurrent discoveries.

	containerizer := turtlefinder.New()
	allnetns := gostwire.Discover(req.Context(), containerizer)

For more information about Gostwire's information model, please see the network
package.

# Linux Capabilities

Please note that the auto-detection of container engines as well as a complete
discovery require the calling process and its OS-level threads to possess
sufficient capabilities. In particular:

  - CAP_SYS_ADMIN
  - CAP_SYS_CHROOT
  - CAP_SYS_PTRACE
  - CAP_DAC_READ_SEARCH
  - CAP_DAC_OVERRIDE
*/
package gostwire
