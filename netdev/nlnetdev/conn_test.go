// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"errors"
	"fmt"

	"github.com/mdlayher/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("netdev NETLINK client", func() {

	BeforeEach(func() {
		oldFamErr := familyErr
		DeferCleanup(func() { familyErr = oldFamErr })
	})

	It("reports an error when family has failed", func() {
		familyErr = fmt.Errorf("generic NETLINK family %q not available, reason: %w",
			NetdevFamilyName, errors.New("foobar"))
		Expect(Dial(nil)).Error().To(MatchError(ContainSubstring("generic NETLINK family")))
	})

	It("reports dial errors", func() {
		familyErr = nil
		Expect(Dial(&netlink.Config{NetNS: 1})).Error().To(HaveOccurred())
	})

})
