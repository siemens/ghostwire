// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package host

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetaHost(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/meta/host package")
}
