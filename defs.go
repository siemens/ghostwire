// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package gostwire

//go:generate go run ./internal/gen/version

// CaptureEnableHeader tells the Ghostwire service to serve its SPA user
// interface with capture button enabled.
const CaptureEnableHeader = "Enable-Monolith"
