// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireXsk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/xsk package")
}
