// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireXskDiscover(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/xsk/discover package")
}
