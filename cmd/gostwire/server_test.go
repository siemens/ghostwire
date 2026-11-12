// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing/fstest"

	"github.com/thediveo/spaserve"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("index.html dynamic variables", func() {

	const index = `<html><script>window.dynvars = {}</script></html>`

	It("inserts brand and disabled capture links", func() {
		rewrite := newDynamicVarsRewriter("Edgeshark", "<svg/>", false)
		req := httptest.NewRequest("GET", "/", nil)
		Expect(rewrite(req, index)).To(Equal(
			`<html><script>window.dynvars = {"enableMonolith":false,"brand":"Edgeshark","brandicon":"\u003csvg/\u003e","preferAppIcons":false}</script></html>`))
	})

	It("inserts the app icons preference", func() {
		rewrite := newDynamicVarsRewriter("Ghostwire", "", true)
		req := httptest.NewRequest("GET", "/", nil)
		Expect(rewrite(req, index)).To(ContainSubstring(`"preferAppIcons":true`))
	})

	It("enables capture links when signalled", func() {
		rewrite := newDynamicVarsRewriter("Ghostwire", "", false)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(gostwire.CaptureEnableHeader, "Affirmative, Dave")
		Expect(rewrite(req, index)).To(ContainSubstring(`"enableMonolith":true`))
	})

	It("doesn't expand submatch references", func() {
		rewrite := newDynamicVarsRewriter("$1 ${2}", "", false)
		req := httptest.NewRequest("GET", "/", nil)
		Expect(rewrite(req, index)).To(ContainSubstring(`"brand":"$1 ${2}"`))
	})

	DescribeTable("passing the base path determined by spaserve",
		func(path string, fwdPrefix string, base string) {
			h := newSPAHandler(fstest.MapFS{
				"index.html": &fstest.MapFile{Data: []byte(
					`<html><base href="/" /><script>window.dynvars = {}</script></html>`)},
			}, "Ghostwire", "", false)
			req := httptest.NewRequest("GET", path, nil)
			if fwdPrefix != "" {
				req.Header.Set(spaserve.ForwardedPrefixHeader, fwdPrefix)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(string(Successful(io.ReadAll(rec.Body)))).To(Equal(
				`<html><base href="` + base + `" /><script>window.dynvars = {"enableMonolith":false,"brand":"Ghostwire","brandicon":"","preferAppIcons":false,"base":"` + base + `"}</script></html>`))
		},
		Entry("directly", "/", "", "/"),
		Entry("directly on a route", "/lochla", "", "/"),
		Entry("behind a path rewriting proxy", "/", "/edgeshark", "/edgeshark/"),
		Entry("on a route behind a path rewriting proxy", "/lochla", "/edgeshark", "/edgeshark/"),
	)

})
