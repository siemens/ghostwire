// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package debug

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDebug(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "afxdp/internal/debug package")
}
