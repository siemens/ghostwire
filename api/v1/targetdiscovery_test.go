// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ohler55/ojg/oj"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("v1 target discovery API", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	It("conforms to its API spec", func() {
		tl := NewTargetDiscoveryResult(disco)
		jtext, err := json.Marshal(tl)
		Expect(err).NotTo(HaveOccurred())

		Expect(validate(v1apispec, "TargetDiscoveryResult", jtext)).To(Succeed())
	})

	It("lists capture targets", func() {
		tl := NewTargetDiscoveryResult(disco)
		jtext, err := json.Marshal(tl)
		Expect(err).NotTo(HaveOccurred())

		v, err := oj.Parse(jtext)
		Expect(err).NotTo(HaveOccurred(), "json: %s", string(jtext))

		Expect(jsnp(v, `$.metadata['creator-id']`).(string)).To(Equal(CreatorID))

		Expect(jsnpsl(v, `$.containers[*]`)).NotTo(BeEmpty())

		Expect(jsnpsl(v, `$.containers[*].name`)).NotTo(ContainElement(BeEmpty()))
		Expect(jsnpsl(v, `$.containers[*].netns`)).NotTo(ContainElement(BeZero()))
		Expect(jsnpsl(v, `$.containers[*].type`)).NotTo(ContainElement(BeZero()))

		Expect(jsnpsl(v, `$.containers[?(@.pid==1 && @.type=='proc')]`)).To(HaveLen(1))

		pods := jsnpsl(v, fmt.Sprintf(`$.containers[?(@.type=='pod' && @.name=='%s')]`, podFQDN))
		Expect(pods).To(HaveLen(1))
		Expect(jsnp(pods[0], `$['network-interfaces']`)).To(ConsistOf("lo", "eth0"))
		Expect(jsnp(pods[0], `$.pid`)).NotTo(BeZero())
		Expect(jsnp(pods[0], `$.starttime`)).NotTo(BeZero())

		bare := jsnpsl(v, fmt.Sprintf(`$.containers[?(@.type=='containerd' && @.name=='%s')]`, bareName))
		Expect(bare).To(HaveLen(1), "%v", v)
		Expect(jsnp(bare[0], `$.prefix`)).To(Equal(barePrefix))
		Expect(jsnp(bare[0], `$['network-interfaces']`)).To(ConsistOf("lo", "eth0"))
	})

})
