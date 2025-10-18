// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/xslog"
	"github.com/thediveo/spaserve"

	gostwire "github.com/siemens/ghostwire/v2"
)

// dynVarsRe matches the window.dynvars assignment, so we can rewrite (or
// rather, insert) the current values of variables that might or might not
// changed based on the particular HTTP request.
var dynVarsRe = regexp.MustCompile(`(<script>window\.dynvars\s*=\s*){}(\s*</script>)`)

var (
	once   sync.Once
	server *http.Server
)

// AddDynamicVars adds, respective inserts, the "dynamic variables" into the
// HTML, CSS and Javascript code (soup) that is the index.html file.
func AddDynamicVars(r *http.Request, index string) string {
	// HAL, do we get signalled to enable capture links? Affirmative, Dave...
	_, enableCaptureLinks := r.Header[gostwire.CaptureEnableHeader]
	dynvars, err := json.Marshal(struct {
		EnableCaptureLinks bool   `json:"enableMonolith"`
		Brand              string `json:"brand"`
		BrandIcon          string `json:"brandicon"`
	}{
		EnableCaptureLinks: enableCaptureLinks,
		Brand:              *brandName,
		BrandIcon:          *brandIcon,
	})
	if err != nil {
		slog.Error("cannot marshal dynamic variables into index.html",
			xslog.Error(err))
		return index
	}
	index = dynVarsRe.ReplaceAllString(string(index), "${1}"+string(dynvars)+"${2}")
	return index
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

func startServer(address string, cmd *cobra.Command, cizer containerizer.Containerizer) (net.Addr, error) {
	// Create the HTTP server listening transport...
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// Finally create the request router and set the routes to the individual
	// handlers.
	r := mux.NewRouter()
	r.Use(requestLogger)
	registerRouteHandlers(r, cmd, cizer)

	r.PathPrefix("/").Handler(spaserve.NewSPAHandler(
		uifs, "index.html", spaserve.WithIndexRewriter(AddDynamicVars)))

	server = &http.Server{Handler: r}
	go func() {
		slog.Info("starting gostwire server",
			slog.String("addr", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			slog.Error("gostwire server failure", xslog.Error(err))
		}
	}()
	return listener.Addr(), nil
}

func stopServer(wait time.Duration) {
	once.Do(func() {
		if server != nil {
			slog.Info("gracefully shutting down gostwire server, waiting...",
				slog.String("maxwait", wait.String()))
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			_ = server.Shutdown(ctx)
			slog.Info("gostwire server stopped.")
		}
	})
}
