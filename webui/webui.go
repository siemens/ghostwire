// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build webui

package webui

import "embed"

//go:embed build
var Webui embed.FS
