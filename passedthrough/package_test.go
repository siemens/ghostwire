// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package passedthrough

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwirePassedthrough(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/passedthrough package")
}
