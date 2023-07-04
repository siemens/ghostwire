// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build webui

package main

import (
	"io/fs"

	"github.com/siemens/ghostwire/v2/webui"
)

var uifs fs.FS

func init() {
	uifs, _ = fs.Sub(webui.Webui, "build")
}
