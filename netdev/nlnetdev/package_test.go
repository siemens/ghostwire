// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireNetdev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/netdev/nlnetdev package")
}
