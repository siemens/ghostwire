// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package brand

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdCliBrand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/cmd/cli/brand package")
}
