// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/xslog"
	"github.com/thediveo/spaserve"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/cmd/cli/brand"
)

// dynVarsRe matches the window.dynvars assignment, so we can rewrite (or
// rather, insert) the current values of variables that might or might not
// changed based on the particular HTTP request.
var dynVarsRe = regexp.MustCompile(`(<script>window\.dynvars\s*=\s*){}(\s*</script>)`)

// newDynamicVarsRewriter returns an index.html rewriter that adds, respective
// inserts, the "dynamic variables" into the HTML, CSS and Javascript code
// (soup) that is the index.html file.
func newDynamicVarsRewriter(brandName, brandIcon string) spaserve.IndexRewriter {
	return func(r *http.Request, index string) string {
		// HAL, do we get signalled to enable capture links? Affirmative, Dave...
		_, enableCaptureLinks := r.Header[gostwire.CaptureEnableHeader]
		dynvars, err := json.Marshal(struct {
			EnableCaptureLinks bool   `json:"enableMonolith"`
			Brand              string `json:"brand"`
			BrandIcon          string `json:"brandicon"`
		}{
			EnableCaptureLinks: enableCaptureLinks,
			Brand:              brandName,
			BrandIcon:          brandIcon,
		})
		if err != nil {
			slog.Error("cannot marshal dynamic variables into index.html",
				xslog.Error(err))
			return index
		}
		return dynVarsRe.ReplaceAllString(index, "${1}"+string(dynvars)+"${2}")
	}
}

// requestLogger is a middleware that closes the specified HTTP handler so that
// requests get logged at info level.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.Info("http request",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path))
		next.ServeHTTP(w, req)
	})
}

// newHandler returns the HTTP request router with the API routes registered via
// plugins, as well as the web UI.
func newHandler(cmd *cobra.Command, cizer containerizer.Containerizer) http.Handler {
	r := mux.NewRouter()
	r.Use(requestLogger)
	registerRouteHandlers(r, cmd, cizer)
	r.PathPrefix("/").Handler(spaserve.NewSPAHandler(
		uifs, "index.html", spaserve.WithIndexRewriter(
			newDynamicVarsRewriter(brand.Name(cmd), brand.Icon(cmd)))))
	return r
}
