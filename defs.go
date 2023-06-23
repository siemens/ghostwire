// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:generate go run ./internal/gen/version
package gostwire

// CaptureEnableHeader tells the Ghostwire service to serve its SPA user
// interface with capture button enabled.
const CaptureEnableHeader = "Enable-Monolith"
