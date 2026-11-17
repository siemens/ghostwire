// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

//go:build !webui

package webui

import (
	"io/fs"
	"os"
)

// FS returns the web UI build from the “webui/build” directory relative to the
// current working directory, as the web UI hasn't been embedded into the
// binary.
func FS() fs.FS {
	return os.DirFS("webui/build")
}
