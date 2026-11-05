// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package initialcgroup

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdCliInitialCgroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/cmd/gostwire/cli/initialcgroup package")
}
