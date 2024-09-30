// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package dur

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireInternalDur(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/internal/dur package")
}
