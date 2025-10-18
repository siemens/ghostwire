// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/xslog"
	"github.com/thediveo/whalewatcher/v2/engineclient/moby"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/cmd/internal/wsconn"
	"github.com/siemens/ghostwire/v2/mobydig"
)

const maxDiggers = 8
const maxVerifiers = 8

var validContainerNamesRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// registerDiscovery registers the /mobydig discovery route and handler with the
// route handler plugin mechanism.
func init() {
	plugger.Group[RouteHandler]().Register(
		func(cmd *cobra.Command, cizer containerizer.Containerizer) (string, string, http.HandlerFunc) {
			return "GET",
				"/mobydig",
				func(w http.ResponseWriter, req *http.Request) {
					ctx := req.Context()
					query := req.URL.Query()
					target := query["target"]
					if len(target) != 1 {
						http.Error(w, "missing or multiple target query parameters", 400)
						return
					}
					targetname := target[0]
					// While limiting target names to at most 255 ASCII
					// characters is arbitrary, it cuts off any big blob query
					// parameter nonsense. Due to DNS label restriction, useful
					// Docker container names are at most 63 characters.
					if !validContainerNamesRegex.MatchString(targetname) || len(targetname) > 255 {
						http.Error(w, "invalid target name", 400)
						return
					}

					conn, err := wsconn.NewWSConn(w, req)
					if err != nil {
						return
					}
					go conn.Watch()
					defer func() {
						conn.Debug("neighborhood nearby scan completed",
							slog.String("target", targetname))
						conn.InitiateGracefulClose(1000, "")
					}()
					conn.Debug("discovering and verifying nearby neighborhood services",
						slog.String("target", targetname))

					allnetns := gostwire.Discover(req.Context(), cizer, nil)
					startContainer := allnetns.Lxkns.Containers.FirstWithNameType(targetname, moby.Type)
					if startContainer == nil {
						slog.Error("Docker container not found",
							slog.String("target", targetname))
						conn.GracefullyClose(4400,
							fmt.Sprintf("Docker container %q not found", targetname))
						return
					}
					conn.Debug("successfully located target container",
						slog.String("target", targetname))
					verdicts, err := mobydig.DigNeighborhoodServices(ctx,
						allnetns.Netns, startContainer,
						maxDiggers, maxVerifiers)
					if err != nil {
						conn.Error("cannot start digging and address verification",
							xslog.Error(err))
						conn.GracefullyClose(4400,
							fmt.Sprintf("cannot start digging and address verification for %q", targetname))
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
