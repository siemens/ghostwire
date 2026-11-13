// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package umem

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireUmem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/xsk/umem package")
}
