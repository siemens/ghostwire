// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package gostwire

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGhostwire(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire package")
}
