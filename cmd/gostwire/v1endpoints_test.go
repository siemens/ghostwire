// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("v1 API endpoints", func() {

	var baseurl string

	BeforeEach(func() {
		addr, _ := startService()
		baseurl = "http://" + addr
	})

	get := func(path string) (int, map[string]any) {
		GinkgoHelper()
		resp := Successful(http.Get(baseurl + path))
		body := Successful(io.ReadAll(resp.Body))
		Expect(resp.Body.Close()).To(Succeed())
		var result map[string]any
		if resp.StatusCode == http.StatusOK {
			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(json.Unmarshal(body, &result)).To(Succeed())
		}
		return resp.StatusCode, result
	}

	It("serves the discovery", func() {
		status, result := get("/json?ieappicons=off")
		Expect(status).To(Equal(http.StatusOK))
		Expect(result).To(HaveKey("metadata"))
		Expect(result).To(HaveKeyWithValue("network-namespaces", Not(BeEmpty())))
	})

	It("serves the capture target discovery", func() {
		status, result := get("/mobyshark")
		Expect(status).To(Equal(http.StatusOK))
		Expect(result).To(HaveKey("metadata"))
		Expect(result).To(HaveKey("containers"))
	})

	DescribeTable("rejects invalid neighborhood targets",
		func(query string) {
			status, _ := get("/mobydig" + query)
			Expect(status).To(Equal(http.StatusBadRequest))
		},
		Entry("missing target", ""),
		Entry("multiple targets", "?target=foo&target=bar"),
		Entry("invalid target name", "?target="+url.QueryEscape("-foo")),
		Entry("overly long target name", "?target="+strings.Repeat("a", 256)),
	)

})
