// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

//go:build webui

package webui

import (
	"embed"
	"io/fs"
)

// Webui contains the embedded web UI build in its “build” directory.
//
//go:embed build
var Webui embed.FS

// FS returns the web UI build embedded into the binary.
func FS() fs.FS {
	uifs, _ := fs.Sub(Webui, "build")
	return uifs
}
