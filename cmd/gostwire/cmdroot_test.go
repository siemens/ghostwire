// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("gostwire command", func() {

	It("shows help with all flags", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(SatisfyAll(
			MatchRegexp(`--http string .*\(default "\[::\]:5000"\)`),
			MatchRegexp(`--shutdown duration .*\(default 15s\)`),
			ContainSubstring("--allowed-origin strings"),
			MatchRegexp(`--brand string .*\(default "Ghostwire"\)`),
			ContainSubstring("--brandicon string"),
			ContainSubstring("--prefer-showing-app-icons"),
			ContainSubstring("--initialcgroup"),
			Not(ContainSubstring("--cgroupswitched")),
			ContainSubstring("--debug"),
			ContainSubstring("--silent"),
			ContainSubstring("--version"),
		))
	})

	It("shows its version", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--version"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(Equal("gostwire version " + gostwire.SemVersion + "\n"))
	})

	It("rejects arguments", func() {
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"foo"})
		rootCmd.SetOut(GinkgoWriter)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(MatchError(ContainSubstring(`unknown command "foo"`)))
	})

	It("fails when it cannot listen on its address", func() {
		occupied := Successful(net.Listen("tcp", "127.0.0.1:0"))
		DeferCleanup(occupied.Close)
		rootCmd := newRootCmd()
		rootCmd.SetContext(context.Background())
		debug.SetWriter(rootCmd, GinkgoWriter)
		rootCmd.SetArgs([]string{"--http", occupied.Addr().String()})
		rootCmd.SetOut(GinkgoWriter)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(MatchError(ContainSubstring("cannot start service")))
	})

	It("serves its API until cancelled", func() {
		addr, _ := startService()
		resp := Successful(http.Get("http://" + addr + "/mobydig"))
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("accepts websocket connections only from allowed origins", func(ctx context.Context) {
		addr, _ := startService("--allowed-origin", "https://gw.example")
		wsurl := "ws://" + addr + "/mobydig?target=non-existing"

		_, resp, err := websocket.DefaultDialer.DialContext(ctx, wsurl,
			http.Header{"Origin": {"https://evil.example"}})
		Expect(err).To(MatchError(websocket.ErrBadHandshake))
		Expect(resp).To(HaveField("StatusCode", http.StatusForbidden))

		conn, resp := Successful2R(websocket.DefaultDialer.DialContext(ctx, wsurl,
			http.Header{"Origin": {"https://gw.example"}}))
		DeferCleanup(resp.Body.Close)
		DeferCleanup(conn.Close)
		_, _, err = conn.ReadMessage()
		Expect(AssignableTo[*websocket.CloseError](err).Code).To(Equal(4400))
	})

})
