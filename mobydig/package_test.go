// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

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

func TestMobyDig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/mobydig package")
}
