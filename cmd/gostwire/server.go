// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

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

// baseHrefRe matches the base element that spaserve has already rewritten to
// the base path of the SPA as seen by the client.
var baseHrefRe = regexp.MustCompile(`<base href="(.*?)"\s*/>`)

// newSPAHandler returns a handler serving the web UI from the specified fs,
// with spaserve rewriting the base element and us adding the dynamic variables.
func newSPAHandler(uifs fs.FS, brandName, brandIcon string) http.Handler {
	return spaserve.NewSPAHandler(uifs, "index.html",
		spaserve.WithIndexRewriter(newDynamicVarsRewriter(brandName, brandIcon)))
}

// newDynamicVarsRewriter returns an index.html rewriter that adds, respective
// inserts, the "dynamic variables" into the HTML, CSS and Javascript code
// (soup) that is the index.html file. This includes the base path of the SPA as
// determined by spaserve from the request and any forwarding proxy headers.
func newDynamicVarsRewriter(brandName, brandIcon string) spaserve.IndexRewriter {
	return func(r *http.Request, index string) string {
		var base string
		if m := baseHrefRe.FindStringSubmatch(index); m != nil {
			base = m[1]
		}
		// HAL, do we get signalled to enable capture links? Affirmative, Dave...
		_, enableCaptureLinks := r.Header[gostwire.CaptureEnableHeader]
		dynvars, err := json.Marshal(struct {
			EnableCaptureLinks bool   `json:"enableMonolith"`
			Brand              string `json:"brand"`
			BrandIcon          string `json:"brandicon"`
			Base               string `json:"base,omitempty"`
		}{
			EnableCaptureLinks: enableCaptureLinks,
			Brand:              brandName,
			BrandIcon:          brandIcon,
			Base:               base,
		})
		if err != nil {
			slog.Error("cannot marshal dynamic variables into index.html",
				xslog.Error(err))
			return index
		}
		// Escape "$" so that it isn't taken as a submatch reference.
		return dynVarsRe.ReplaceAllString(index,
			"${1}"+strings.ReplaceAll(string(dynvars), "$", "$$")+"${2}")
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
	r.PathPrefix("/").Handler(newSPAHandler(uifs, brand.Name(cmd), brand.Icon(cmd)))
	return r
}
