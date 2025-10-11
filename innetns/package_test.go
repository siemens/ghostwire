// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package innetns

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireInNetns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/innetns package")
}
