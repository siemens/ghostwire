// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecorator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator/ieappicon package")
}
