// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package appicons

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdCliAppIcons(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/cmd/gostwire/cli/appicons package")
}
