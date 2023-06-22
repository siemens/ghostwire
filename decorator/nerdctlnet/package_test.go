// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctlnet

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecoratorNerdctlnet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator/nerdctlnet package")
}
