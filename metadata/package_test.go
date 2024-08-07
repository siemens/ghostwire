// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package metadata

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetadata(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/metadata package")
}
