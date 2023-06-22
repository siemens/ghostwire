// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"

	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/mobydig/dig"
	"github.com/siemens/mobydig/verifier"
)

// JSONTextualRepresentation is JSON data in its textual string "storage"
// format.
type JSONTextualRepresentation string

// String returns the JSON as a string. Without a Stringer Gomega's MatchJSON
// will fail as this is a new type (albeit based on string) and without it
// explicitly giving a Stringer, MatchJSON won't ever match.
func (j JSONTextualRepresentation) String() string {
	return string(j)
}

// FQDNAddressVerdict represents the outcome of digging an FQDN and then
// checking the dug IP addresses for reachability. FQDNAddressVerdict can
// represent intermediate steps, so they're not necessarily reporting final
// verdicts only.
type FQDNAddressVerdict struct {
	FQDN    string `json:"fqdn,omitempty"`
	Address string `json:"address,omitempty"`
	Quality string `json:"quality,omitempty"`
	Err     string `json:"error,omitempty"`
}

// DigNeighborhoodServices discovers the neighborhood Docker services
// addressable using DNS service and container names and then digs their IP
// addresses and pings them. The ongoing intermediate results as well as the
// final verdicts are then streamed to the returned channel in JSON textual
// format.
func DigNeighborhoodServices(
	ctx context.Context,
	m network.NetworkNamespaces,
	startContainer *model.Container,
	maxDiggers int,
	maxPingers int,
) (<-chan JSONTextualRepresentation, error) {
	services, err := NeighborhoodServices(m, startContainer)
	if err != nil {
		return nil, err
	}
	netnsref := startContainer.Process.Namespaces[model.NetNS].Ref()
	if len(netnsref) != 1 {
		log.Errorf("invalid network namespace reference for Docker container %q: %v",
			startContainer.Name, netnsref)
		return nil, fmt.Errorf("invalid network namespace reference for Docker container %q",
			startContainer.Name)
	}
	// Now put the required processing elements and their plumbing in place.
	//
	//   - Digger producing IP addresses from a list of FQDNs.
	//   - Verifier consuming the IPs and checking them, producing "verdicts".
	digger, diggernews, err := dig.New(maxDiggers, netnsref[0])
	if err != nil {
		return nil, fmt.Errorf("cannot dig address information: %w", err)
	}
	verifier, news := verifier.New(maxPingers, netnsref[0])
	go verifier.Verify(ctx, diggernews)
	// Then feed the information about attached networks and their names into
	// the Digger, so they can be processed and move through the different
	// stages. Then close the input stream and wait for all the data to pass the
	// stages and finally get rendered a last time.
	fqdns := []string{}
	for _, service := range services {
		fqdns = append(fqdns, service.DnsNames()...)
	}
	go func() {
		digger.DigFQDNs(ctx, fqdns)
		// wait for all digging to be done and then closes the inter-stage
		// channel connection between the digger and the verifier, thus telling
		// the verifier that there won't be more work for today after draining
		// the inter-stage channel.
		digger.StopWait()
	}()
	// Pull off the address resolution and validation (interim) verdicts and
	// pass them on as JSON in textual representation.
	verdictCh := make(chan JSONTextualRepresentation)
	var closeCh sync.Once
	go func() {
		for {
			select {
			case namedAddr, ok := <-news:
				if !ok {
					closeCh.Do(func() { close(verdictCh) })
					return
				}
				qa := namedAddr.QA()
				errstr := ""
				if err := namedAddr.Err(); err != nil {
					errstr = err.Error()
				}
				log.Debugf("fqdn: %q, IP: %s, quality: %s, err: %q",
					namedAddr.Name(), qa.Address, qa.Quality.String(), errstr)
				jtext, _ := json.Marshal(FQDNAddressVerdict{
					FQDN:    namedAddr.Name(),
					Address: qa.Address,
					Quality: qa.Quality.String(),
					Err:     errstr,
				})
				verdictCh <- JSONTextualRepresentation(jtext)
			case <-ctx.Done():
				closeCh.Do(func() { close(verdictCh) })
				return
			}
		}
	}()
	return verdictCh, nil
}
