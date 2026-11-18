// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"io"
	"net/http"
	"testing/fstest"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/safe"

	gostwire "github.com/siemens/ghostwire/v2"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("composing the service", func() {

	get := func(url string) (int, string) {
		GinkgoHelper()
		resp := Successful(http.Get(url))
		body := Successful(io.ReadAll(resp.Body))
		Expect(resp.Body.Close()).To(Succeed())
		return resp.StatusCode, string(body)
	}

	DescribeTable("using its name, description, and version",
		func(arg string, expected types.GomegaMatcher) {
			var out safe.Buffer
			rootCmd := NewRootCmd(
				WithName("foobar"),
				WithDescription("foobar discovery service"),
				WithVersion("6.6.6"))
			rootCmd.SetArgs([]string{arg})
			rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
			Expect(rootCmd.Execute()).To(Succeed())
			Expect(out.String()).To(expected)
		},
		Entry("help", "--help", SatisfyAll(
			HavePrefix("foobar discovery service\n"),
			ContainSubstring("foobar [flags]"),
			MatchRegexp(`--http string .*\(default "\[::\]:5000"\)`),
		)),
		Entry("version", "--version", Equal("foobar version 6.6.6\n")),
	)

	It("logs its name and versions", func() {
		_, logs := startServiceWith([]Option{
			WithName("foobar"),
			WithDescription("foobar discovery service"),
			WithVersion("6.6.6"),
		})
		Expect(logs.String()).To(SatisfyAll(
			ContainSubstring(`msg="foobar discovery service" version=6.6.6 ghostwire=`+gostwire.SemVersion),
			ContainSubstring(`msg="starting foobar server"`),
		))
	})

	It("doesn't log the Ghostwire version separately by default", func() {
		_, logs := startService()
		Expect(logs.String()).To(SatisfyAll(
			ContainSubstring(`msg="Gostwire \"The Sequel\" virtual network topology and configuration discovery service" version=`+gostwire.SemVersion+"\n"),
			ContainSubstring(`msg="starting gostwire server"`),
		))
	})

	It("defaults to a different brand", func() {
		var out safe.Buffer
		rootCmd := NewRootCmd(WithBrand("Foobar"))
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(`--brand string .*\(default "Foobar"\)`))

		addr, _ := startServiceWith([]Option{
			WithBrand("Foobar"),
			WithUI(fstest.MapFS{
				"index.html": &fstest.MapFile{Data: []byte(`<script>window.dynvars = {}</script>`)},
			}),
		})
		_, body := get("http://" + addr + "/")
		Expect(body).To(ContainSubstring(`"brand":"Foobar"`))
	})

	It("serves no web UI unless passed one", func() {
		addr, _ := startService()
		status, _ := get("http://" + addr + "/w")
		Expect(status).To(Equal(http.StatusNotFound))
	})

	It("serves a different UI", func() {
		addr, _ := startServiceWith([]Option{
			WithUI(fstest.MapFS{
				"index.html": &fstest.MapFile{Data: []byte(
					`<html><script>window.dynvars = {}</script>foobar</html>`)},
			}),
		}, "--brand", "Foobar")
		status, body := get("http://" + addr + "/w")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(SatisfyAll(
			ContainSubstring(`"brand":"Foobar"`),
			ContainSubstring("foobar</html>"),
		))
	})

	It("serves additional routes, but without shadowing", func() {
		// set before the service logs its address, so no data race.
		var flagged string
		var gotCizer containerizer.Containerizer
		addr, _ := startServiceWith([]Option{
			WithUI(fstest.MapFS{
				"index.html": &fstest.MapFile{Data: []byte(`<html>UI</html>`)},
			}),
			WithRoutes(func(r *mux.Router, cmd *cobra.Command, cizer containerizer.Containerizer) {
				gotCizer = cizer
				flagged, _ = cmd.Flags().GetString("brand")
				r.HandleFunc("/api/v1/foobar", func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("foobar"))
				}).Methods(http.MethodGet)
				r.HandleFunc("/mobydig", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}).Methods(http.MethodGet)
			}),
			WithRoutes(nil),
		}, "--brand", "Foobar")
		Expect(flagged).To(Equal("Foobar"))
		Expect(gotCizer).NotTo(BeNil())

		status, body := get("http://" + addr + "/api/v1/foobar")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("foobar"))

		status, _ = get("http://" + addr + "/mobydig")
		Expect(status).To(Equal(http.StatusBadRequest))

		status, body = get("http://" + addr + "/api/v1/barfoo")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("UI"))
	})

	Context("overriding and removing routes", func() {

		ui := WithUI(fstest.MapFS{
			"index.html": &fstest.MapFile{Data: []byte(`<html>UI</html>`)},
		})

		It("overrides a route, optionally calling the upstream handler", func() {
			// set before the service logs its address, so no data race.
			var gotCmd *cobra.Command
			var gotCizer containerizer.Containerizer
			addr, _ := startServiceWith([]Option{
				ui,
				WithRouteOverride(RoutePluginMobydig, func(upstream http.HandlerFunc, cmd *cobra.Command, cizer containerizer.Containerizer) http.HandlerFunc {
					gotCmd, gotCizer = cmd, cizer
					return func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("X-Override", "inner")
						if req.URL.Query().Has("target") {
							upstream(w, req)
							return
						}
						w.WriteHeader(http.StatusTeapot)
					}
				}),
				WithRouteOverride(RoutePluginMobydig, func(upstream http.HandlerFunc, _ *cobra.Command, _ containerizer.Containerizer) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("X-Override", "outer")
						upstream(w, req)
					}
				}),
				WithRouteOverride(RoutePluginMobydig, nil),
			})
			Expect(gotCmd).NotTo(BeNil())
			Expect(gotCizer).NotTo(BeNil())

			resp := Successful(http.Get("http://" + addr + "/mobydig"))
			Expect(resp.Body.Close()).To(Succeed())
			Expect(resp.StatusCode).To(Equal(http.StatusTeapot))
			Expect(resp.Header.Values("X-Override")).To(Equal([]string{"outer", "inner"}))

			status, _ := get("http://" + addr + "/mobydig?target=-foo")
			Expect(status).To(Equal(http.StatusBadRequest))
		})

		It("removes a route, falling back to the UI", func() {
			addr, _ := startServiceWith([]Option{ui, WithoutRoutes(RoutePluginMobydig)})
			status, body := get("http://" + addr + "/mobydig")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("<html>UI</html>"))
		})

		It("removes a route, so that it can be reimplemented", func() {
			addr, _ := startServiceWith([]Option{
				ui,
				WithoutRoutes(RoutePluginMobydig),
				WithRoutes(func(r *mux.Router, _ *cobra.Command, _ containerizer.Containerizer) {
					r.HandleFunc("/mobydig", func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte("reimplemented"))
					}).Methods(http.MethodGet)
				}),
			})
			status, body := get("http://" + addr + "/mobydig")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("reimplemented"))
		})

		DescribeTable("refusing to start with invalid route plugin names",
			func(opts []Option, expected string) {
				rootCmd := NewRootCmd(opts...)
				rootCmd.SetContext(context.Background())
				rootCmd.SetArgs([]string{"--http", "127.0.0.1:0"})
				rootCmd.SetOut(GinkgoWriter)
				rootCmd.SetErr(GinkgoWriter)
				Expect(rootCmd.Execute()).To(MatchError(ContainSubstring(expected)))
			},
			Entry("unknown overridden route",
				[]Option{WithRouteOverride("foo", func(upstream http.HandlerFunc, _ *cobra.Command, _ containerizer.Containerizer) http.HandlerFunc {
					return upstream
				})},
				`cannot override unknown route plugin "foo", known route plugins: json, mobydig, mobyshark`),
			Entry("unknown removed route",
				[]Option{WithoutRoutes(RoutePluginJSON, "bar")},
				`cannot remove unknown route plugin "bar"`),
			Entry("both overridden and removed route",
				[]Option{
					WithRouteOverride(RoutePluginJSON, func(upstream http.HandlerFunc, _ *cobra.Command, _ containerizer.Containerizer) http.HandlerFunc {
						return upstream
					}),
					WithoutRoutes(RoutePluginJSON),
				},
				`cannot both override and remove route plugin "json"`),
		)

	})

})
