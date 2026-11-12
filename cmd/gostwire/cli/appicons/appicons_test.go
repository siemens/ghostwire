// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package appicons

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("app icons flag", func() {

	It("registers as a CLI setup plugin", func() {
		Expect(plugger.Group[cliplugin.SetupCLI]().Plugins()).To(ContainElement("ghostwire/appicons"))
	})

	DescribeTable("getting the app icons preference",
		func(args []string, expected bool) {
			var actual bool
			root := &cobra.Command{
				Use: "root",
				Run: func(cmd *cobra.Command, _ []string) {
					actual = PreferShowing(cmd)
				},
			}
			root.SetOut(GinkgoWriter)
			root.SetErr(GinkgoWriter)
			setupCLI(root)
			root.SetArgs(append([]string{}, args...))
			Expect(root.Execute()).To(Succeed())
			Expect(actual).To(Equal(expected))
		},
		Entry("default", nil, false),
		Entry("preferred", []string{"--" + PreferShowingAppIconsFlag}, true),
		Entry("explicitly not preferred", []string{"--" + PreferShowingAppIconsFlag + "=false"}, false),
	)

})
