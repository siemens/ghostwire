// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/siemens/ghostwire/v2/cmd/gostwire/app"
	"github.com/siemens/ghostwire/v2/webui"
)

func main() {
	app.Main(app.WithUI(webui.FS()))
}
