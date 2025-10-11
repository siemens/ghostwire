// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package getstdout

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetStdout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "afxdp/cmd/internal/getstdout")
}
