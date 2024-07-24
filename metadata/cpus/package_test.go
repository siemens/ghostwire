// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cpus

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetadataCPUs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/metadata/cpus package")
}
