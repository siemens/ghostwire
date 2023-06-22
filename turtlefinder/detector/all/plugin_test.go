// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package all

import (
	"github.com/siemens/ghostwire/v2/turtlefinder/detector"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/go-plugger/v3"
)

var _ = Describe("detector plugins", func() {

	It("all register the Detector plugin interface and return engine process names", func() {
		namers := plugger.Group[detector.Detector]().Symbols()
		Expect(namers).To(HaveLen(2))
		names := []string{}
		for _, namer := range namers {
			names = append(names, namer.EngineNames()...)
		}
		Expect(names).To(ConsistOf(
			"containerd", "dockerd",
		))
	})

})
