// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build !webui

package main

import "os"

var uifs = os.DirFS("webui/build")
