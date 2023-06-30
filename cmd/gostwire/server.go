// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/mux"
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/spaserve"
)

// dynVarsRe matches the window.dynvars assignment, so we can rewrite (or
// rather, insert) the current values of variables that might or might not
// changed based on the particular HTTP request.
var dynVarsRe = regexp.MustCompile(`(<script>window\.dynvars=){}(</script>)`)

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
		log.Errorf("cannot marshal dynamic variables into index.html, reason: %s",
			err.Error())
		return index
	}
	index = dynVarsRe.ReplaceAllString(string(index), "${1}"+string(dynvars)+"${2}")
	return index
}

// requestLogger is a middleware that closes the specified HTTP handler so that
// requests get logged at info level.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Infof("http %s %s", req.Method, req.RequestURI)
		next.ServeHTTP(w, req)
	})
}

func startServer(address string, cizer containerizer.Containerizer) (net.Addr, error) {
	// Create the HTTP server listening transport...
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// Finally create the request router and set the routes to the individual
	// handlers.
	r := mux.NewRouter()
	r.Use(requestLogger)
	registerDiscovery(cizer)
	registerMobyDigger(cizer)
	registerRouteHandlers(r)

	r.PathPrefix("/").Handler(spaserve.NewSPAHandler(
		os.DirFS("webui/build"), "index.html", spaserve.WithIndexRewriter(AddDynamicVars)))

	server = &http.Server{Handler: r}
	go func() {
		log.Infof("starting gostwire server to serve at %s", listener.Addr().String())
		if err := server.Serve(listener); err != nil {
			log.Errorf("gostwire server error: %s", err.Error())
		}
	}()
	return listener.Addr(), nil
}

func stopServer(wait time.Duration) {
	once.Do(func() {
		if server != nil {
			log.Infof("gracefully shutting down gostwire server, waiting up to %s...",
				wait)
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			_ = server.Shutdown(ctx)
			log.Infof("gostwire server stopped.")
		}
	})
}
