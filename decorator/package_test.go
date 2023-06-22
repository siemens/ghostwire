// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package decorator

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecorator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator package")
}
