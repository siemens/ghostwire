// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"net/http/httptest"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("index.html dynamic variables", func() {

	const index = `<html><script>window.dynvars = {}</script></html>`

	It("inserts brand and disabled capture links", func() {
		rewrite := newDynamicVarsRewriter("Edgeshark", "<svg/>")
		req := httptest.NewRequest("GET", "/", nil)
		Expect(rewrite(req, index)).To(Equal(
			`<html><script>window.dynvars = {"enableMonolith":false,"brand":"Edgeshark","brandicon":"\u003csvg/\u003e"}</script></html>`))
	})

	It("enables capture links when signalled", func() {
		rewrite := newDynamicVarsRewriter("Ghostwire", "")
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(gostwire.CaptureEnableHeader, "Affirmative, Dave")
		Expect(rewrite(req, index)).To(ContainSubstring(`"enableMonolith":true`))
	})

})
