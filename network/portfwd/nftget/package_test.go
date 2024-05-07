// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func init() {
	// avoid M0 ending up wedged as it was used during a throw-away namespace
	// switch, but as M0 is special it cannot be killed.
	runtime.LockOSThread()
}

func TestGostwireNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/network/portfwd/nftget package")
}
