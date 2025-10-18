// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/xslog"

	gostwire "github.com/siemens/ghostwire/v2"
	apiv1 "github.com/siemens/ghostwire/v2/api/v1"
	"github.com/siemens/ghostwire/v2/decorator/ieappicon"
)

// registerDiscovery registers the /json and /mobyshark discovery routes and
// handlers with the route handler plugin mechanism.
func init() {
	plugger.Group[RouteHandler]().Register(
		func(cmd *cobra.Command, cizer containerizer.Containerizer) (string, string, http.HandlerFunc) {
			return "GET",
				"/json",
				func(w http.ResponseWriter, req *http.Request) {
					discoveryLabels := map[string]string{}
					query := req.URL.Query()
					if ieappicons, ok := query["ieappicons"]; ok {
						discoveryLabels[ieappicon.IEAppDiscoveryLabel] = ieappicons[0]
					}
					allnetns := gostwire.Discover(req.Context(), cizer, discoveryLabels)
					result := apiv1.NewDiscoveryResult(allnetns)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(&result)
					if err != nil {
						slog.Error("discovery result marshalling failure",
							xslog.Error(err))
					}
				}
		}, plugger.WithPlugin("json"))
	plugger.Group[RouteHandler]().Register(
		func(cmd *cobra.Command, cizer containerizer.Containerizer) (string, string, http.HandlerFunc) {
			return "GET",
				"/mobyshark",
				func(w http.ResponseWriter, req *http.Request) {
					allnetns := gostwire.Discover(req.Context(), cizer, nil)
					result := apiv1.NewTargetDiscoveryResult(allnetns)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(&result)
					if err != nil {
						slog.Error("capture target discovery result marshalling failure",
							xslog.Error(err))
					}
				}
		}, plugger.WithPlugin("mobyshark"))
}
