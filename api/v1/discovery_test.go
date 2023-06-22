// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"os"

	"github.com/ohler55/ojg/oj"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
)

var _ = Describe("v1 discovery API", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	It("conforms to its API spec", func() {
		d := NewDiscoveryResult(disco)
		jtext, err := json.Marshal(&d)
		Expect(err).NotTo(HaveOccurred())

		defer func(ml int) { format.MaxLength = ml }(format.MaxLength)
		Expect(validate(v1apispec, "DiscoveryResult", jtext)).To(Succeed())
	})

	It("contains discovery information", func() {
		d := NewDiscoveryResult(disco)
		jtext, err := json.Marshal(&d)
		Expect(err).NotTo(HaveOccurred())

		v, err := oj.Parse(jtext)
		Expect(err).NotTo(HaveOccurred(), "json: %s", string(jtext))

		Expect(jsnp(v, `$.metadata['creator-id']`).(string)).To(Equal(CreatorID))
		Expect(jsnp(v, `$['network-namespaces']`)).To(HaveLen(len(disco.Netns)))

		initialpidnsid, err := ops.NamespacePath("/proc/1/ns/pid").ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(jsnp(v, `$['pid-namespaces']`)).To(HaveLen(1))
		Expect(jsnp(v, `$['pid-namespaces'][0].children`)).To(
			HaveLen(len(disco.Lxkns.Namespaces[model.PIDNS][initialpidnsid].(model.Hierarchy).Children())))
		Expect(jsnpsl(v, `$['pid-namespaces']..pidnsid`)).NotTo(
			Or(BeEmpty(), ContainElement(BeZero())))
	})

})
