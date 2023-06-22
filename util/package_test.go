// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package util

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/Util package")
}
