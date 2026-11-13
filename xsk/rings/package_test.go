// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rings

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireLotR(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/xsk/rings package")
}
