// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecoratorDockernet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator/dockernet package")
}
