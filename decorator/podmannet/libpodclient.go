// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package podmannet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

// UserAgent specifies the HTTP agent string used when talking to podman's
// libpod API.
const UserAgent = "Gostwire (The Sequel)"

// Client is a minimalist HTTP-over-UDS (unix domain socket) client for
// conversing with podmen libpod API endpoints.
type Client struct {
	httpClient    *http.Client
	endpointURL   *url.URL
	libpodVersion string // libpod API semver, without "v" prefix.
}

// newLibpodClient returns a new podman libpod API client. The endpoint must be
// using the "unix" protocol.
//
// Please note that this libpod API client is absolutely minimalist and just
// suffices for querying the podman-managed networks.
func newLibpodClient(endpoint string, libpodapiversion string) (*Client, error) {
	epurl, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint, reason: %w", err)
	}
	if epurl.Scheme != "unix" {
		return nil, fmt.Errorf("unsupported endpoint protocol '%s'", epurl.Scheme)
	}
	c := &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		},
		endpointURL:   epurl,
		libpodVersion: libpodapiversion,
	}
	dialer := &net.Dialer{
		// same as Docker's unix socket default transport configuration, see
		// also
		// https://github.com/docker/go-connections/blob/fa09c952e3eadbffaf8afc5b8a1667158ba38ace/sockets/sockets.go#L11
		Timeout: 10 * time.Second,
	}
	c.httpClient.Transport.(*http.Transport).DialContext = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		// we don't want to dial the libpod API endpoint, but instead the engine
		// API endpoint as such...
		return dialer.DialContext(ctx, epurl.Scheme, epurl.Path)
	}
	return c, nil
}

// Close closes idle connections.
func (c *Client) Close() error {
	if c.httpClient == nil {
		return nil
	}
	c.httpClient.CloseIdleConnections()
	return nil
}

// apiPath takes a non-versioned libpod API endpoint, such as “/info” and
// “networks/json”; it then returns a versioned libpod path when the libpod
// version is already known, such as “/v1.2.3/libpod/networks/json”. Otherwise
// it returns a “/v0/libpod/...”-based path. In consequence, without the
// libpodVersion set on the Client, only use the “/info” service endpoint, as
// this seems to be version-independent, but still needs any version in its
// endpoint path.
func (c *Client) apiPath(apipath string) string {
	if c.libpodVersion == "" {
		// use only for initial libpod info (API version) retrieval; please note
		// that all libpod API endpoints are versioned, there are not
		// un-versioned endpoints like the Docker API does.
		return path.Join("/v0/libpod", apipath)
	}
	return path.Join("/v"+c.libpodVersion+"/libpod", apipath)
}

// get issues an HTTP GET request for the specified (yet unversioned) API
// endpoint, such as “/networks/json”. It then returns the HTTP response or an
// error.
func (c *Client) get(ctx context.Context, apipath string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://localhost"+c.apiPath(apipath),
		nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("podman service returned status code %d", resp.StatusCode)
	}
	return resp, nil
}

// ensureReaderClosed helper to drain any service response.
func ensureReaderClosed(resp *http.Response) {
	if resp.Body == nil {
		return
	}
	_, _ = io.CopyN(io.Discard, resp.Body, 512)
	resp.Body.Close()
}

// NetworkResource grabs just the few things from a podman network we're
// interested here for the purposes of correctly decorating network interfaces
// with podman network names. We simply ignore all the other JSON salad returned
// from the “/vX/libpod/networks/json” endpoint.
type NetworkResource struct {
	Name             string            `json:"name"`              // name of the network
	ID               string            `json:"id"`                // unique ID within the particular podman engine instance
	Driver           string            `json:"driver"`            // name of the driver; "bridge", "macvlan", "ipvlan"
	NetworkInterface string            `json:"network_interface"` // name of the associated (master) network interface
	Internal         bool              `json:"internal"`          // network is host-internal only, without external connectivity
	Labels           map[string]string `json:"labels"`
}

// networkList returns the list of managed podman networks.
func (c *Client) networkList(ctx context.Context) ([]NetworkResource, error) {
	var netrscs []NetworkResource
	resp, err := c.get(ctx, "/networks/json")
	defer ensureReaderClosed(resp)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&netrscs)
	return netrscs, err
}
