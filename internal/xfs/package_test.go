// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package xfs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireInternalXfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/internal/xfs package")
}
