// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"net/http"

	gostwire "github.com/siemens/ghostwire/v2"
	apiv1 "github.com/siemens/ghostwire/v2/api/v1"
	"github.com/siemens/ghostwire/v2/decorator/ieappicon"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
)

// registerDiscovery registers the /json discovery route and handler with the
// route handler plugin mechanism.
func registerDiscovery(cizer containerizer.Containerizer) {
	plugger.Group[RouteHandler]().Register(
		func() (string, string, http.HandlerFunc) {
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
						log.Errorf("discovery result marshalling error: %s", err.Error())
					}
				}
		}, plugger.WithPlugin("json"))
	plugger.Group[RouteHandler]().Register(
		func() (string, string, http.HandlerFunc) {
			return "GET",
				"/mobyshark",
				func(w http.ResponseWriter, req *http.Request) {
					allnetns := gostwire.Discover(req.Context(), cizer, nil)
					result := apiv1.NewTargetDiscoveryResult(allnetns)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(&result)
					if err != nil {
						log.Errorf("capture target discovery result marshalling error: %s", err.Error())
					}
				}
		}, plugger.WithPlugin("mobyshark"))
}
