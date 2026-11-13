// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireSysfsguess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/netdev/sysfsguess package")
}
