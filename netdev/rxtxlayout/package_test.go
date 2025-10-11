// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TesGostwireRxtxlayout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/netdev/rxtxlayout package")
}
