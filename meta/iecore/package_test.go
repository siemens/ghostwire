// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package iecore

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireMetaIecore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/meta/engines package")
}
