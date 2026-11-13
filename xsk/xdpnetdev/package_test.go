// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xdpnetdev

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireXdpnetdev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/xsk/xdpnetdev package")
}
