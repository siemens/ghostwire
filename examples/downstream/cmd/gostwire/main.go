// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/siemens/ghostwire/v2/cmd/gostwire/app"
	"github.com/siemens/ghostwire/v2/webui"

	_ "example.org/ghostwire-downstream/decorator/hello" // additional decorator
	"example.org/ghostwire-downstream/endpoint"
)

// Version gets stamped at build time using -ldflags "-X main.Version=...".
var Version = "0.0.0-dev"

func main() {
	app.Main(
		app.WithName("gostwire-example"),
		app.WithDescription("example downstream build of the Gostwire discovery service"),
		app.WithVersion(Version),
		app.WithUI(webui.FS()),
		app.WithRoutes(endpoint.Register),
	)
}
