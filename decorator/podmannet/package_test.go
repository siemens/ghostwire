// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package podmannet

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecoratorPodmannet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator/podmannet package")
}
