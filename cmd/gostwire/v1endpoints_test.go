// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/net"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"

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

	It("streams the neighborhood services of a Docker container", SpecTimeout(60*time.Second), func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("test.gostwire.mobydig=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })
		const name = "test-gostwire-mobydig"
		// only custom networks come with Docker's embedded DNS resolver.
		testnet := Successful(sess.CreateNetwork(ctx, name, net.WithInternal()))
		_ = Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(name),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithNetwork(testnet.ID)))

		wsurl := "ws" + strings.TrimPrefix(baseurl, "http") + "/mobydig?target=" + name
		conn, resp, err := websocket.DefaultDialer.DialContext(ctx, wsurl, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		DeferCleanup(conn.Close)
		Expect(conn.SetReadDeadline(time.Now().Add(30 * time.Second))).To(Succeed())
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				Expect(websocket.IsCloseError(err, websocket.CloseNormalClosure)).To(BeTrue(), "%v", err)
				break
			}
		}
	})

})
