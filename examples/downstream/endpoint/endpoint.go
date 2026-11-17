// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package endpoint

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/xslog"
)

// Hello is the response of the “/api/v1/hello” endpoint.
type Hello struct {
	Greeting   string `json:"greeting"`
	Containers int    `json:"containers"`
}

// Register registers the “/api/v1/hello” endpoint; it is an app.RoutesFunc.
func Register(r *mux.Router, _ *cobra.Command, cizer containerizer.Containerizer) {
	r.HandleFunc("/api/v1/hello", func(w http.ResponseWriter, req *http.Request) {
		result := gostwire.Discover(req.Context(), cizer, nil)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Hello{
			Greeting:   "hello",
			Containers: len(result.Lxkns.Containers),
		}); err != nil {
			slog.Error("cannot encode hello response", xslog.Error(err))
		}
	}).Methods(http.MethodGet)
}
