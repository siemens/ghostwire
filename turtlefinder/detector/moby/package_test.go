// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package moby

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireTurtlefinderDetectorMoby(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/turtlefinder/detector/moby package")
}
