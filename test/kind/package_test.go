// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build kind
// +build kind

package kind

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/test/kind package")
}
