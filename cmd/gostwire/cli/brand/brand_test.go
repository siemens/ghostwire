// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package brand

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("brand flags", func() {

	It("registers as a CLI setup plugin", func() {
		Expect(plugger.Group[cliplugin.SetupCLI]().Plugins()).To(ContainElement("ghostwire/brand"))
	})

	DescribeTable("getting the brand",
		func(args []string, name string, icon string) {
			var actualName, actualIcon string
			root := &cobra.Command{
				Use: "root",
				Run: func(cmd *cobra.Command, _ []string) {
					actualName = Name(cmd)
					actualIcon = Icon(cmd)
				},
			}
			root.SetOut(GinkgoWriter)
			root.SetErr(GinkgoWriter)
			setupCLI(root)
			root.SetArgs(append([]string{}, args...))
			Expect(root.Execute()).To(Succeed())
			Expect(actualName).To(Equal(name))
			Expect(actualIcon).To(Equal(icon))
		},
		Entry("defaults", nil, DefaultBrand, ""),
		Entry("plain icon",
			[]string{"--" + BrandFlag, "Edgeshark", "--" + BrandIconFlag, "<svg/>"},
			"Edgeshark", "<svg/>"),
		Entry("base64 icon",
			[]string{"--" + BrandIconFlag, base64.StdEncoding.EncodeToString([]byte("<svg/>"))},
			DefaultBrand, "<svg/>"),
	)

})
