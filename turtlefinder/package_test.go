// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireTurtlefinder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/turtlefinder package")
}
