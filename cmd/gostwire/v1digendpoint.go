// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"net/http"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/cmd/internal/wsconn"
	"github.com/siemens/ghostwire/v2/mobydig"

	"github.com/gorilla/websocket"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

const maxDiggers = 8
const maxVerifiers = 8

// registerDiscovery registers the /json discovery route and handler with the
// route handler plugin mechanism.
func registerMobyDigger(cizer containerizer.Containerizer) {
	plugger.Group[RouteHandler]().Register(
		func() (string, string, http.HandlerFunc) {
			return "GET",
				"/mobydig",
				func(w http.ResponseWriter, req *http.Request) {
					ctx := req.Context()
					query := req.URL.Query()
					target := query["target"]
					if len(target) != 1 {
						http.Error(w, "missing or multiple target query parameters", 400)
					}

					conn, err := wsconn.NewWSConn(w, req)
					if err != nil {
						return
					}
					go conn.Watch()
					defer func() {
						conn.Debugf("neighborhood scan nearby %q complete", target[0])
						conn.InitiateGracefulClose(1000, "")
					}()
					conn.Debugf("discovering and verifying nearby neighborhood services at %q", target[0])

					allnetns := gostwire.Discover(req.Context(), cizer, nil)
					startContainer := allnetns.Lxkns.Containers.FirstWithNameType(target[0], moby.Type)
					if startContainer == nil {
						log.Errorf("Docker container %q not found", target[0])
						conn.GracefullyClose(4400,
							fmt.Sprintf("Docker container %q not found", target[0]))
						return
					}
					conn.Debugf("successfully located %q", target[0])
					verdicts, err := mobydig.DigNeighborhoodServices(ctx,
						allnetns.Netns, startContainer,
						maxDiggers, maxVerifiers)
					if err != nil {
						conn.Errorf("cannot start digging and address verification: %s", err.Error())
						conn.GracefullyClose(4400,
							fmt.Sprintf("cannot start digging and address verification for %q", target[0]))
						return
					}
					for {
						select {
						case verdict, ok := <-verdicts:
							if !ok {
								return
							}
							conn.Conn.WriteMessage(websocket.TextMessage, []byte(verdict))
						case <-ctx.Done():
							return
						}
					}
				}
		}, plugger.WithPlugin("mobydig"))
}
