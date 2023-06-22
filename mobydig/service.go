// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

// Services describes the services and their containers that are DNS-addressable
// from the point of a particular container.
type Services []ServiceOnNetworks

// ServiceOnNetworks describes a Docker composer service (or alternatively a
// single, lonely stand-alone container) reachable on one or more Docker
// networks from the perspective of a particular container.
type ServiceOnNetworks struct {
	Name          string     // service name DNS label for a group of containers; zero for non-service containers
	Servants      Containers // the containers providing the service
	NetworkLabels []string   // network name(s) doubling as (single) DNS (tld) label
}

// Deduplicate a list of services, based on the service names *AND* the
// containers powering the service. The service names in themselves might not
// necessarily "globally" unique within a Docker host. As stand-alone containers
// are also DNS-addressable we maintain pseudo Service instances for them, but
// with an empty Service.Name.
func (s Services) Deduplicate() Services {
	// index services by their names, with stand-alone containers here using
	// pseudo service names in the form of ":$CONTAINER.NAME".
	uniques := map[string]ServiceOnNetworks{}
	for _, service := range s {
		servicename := service.Name
		if servicename == "" {
			if len(service.Servants) == 0 {
				continue // skip incomplete information
			}
			servicename = ":" + service.Servants[0].Name
		}
		if uniqserv, ok := uniques[servicename]; ok {
			uniqserv.Servants = uniqserv.Servants.Merge(service.Servants)
		andNowForSomethingCompletelyDifferent:
			for _, anotherNetLabel := range service.NetworkLabels {
				for _, uniqservsNetLabel := range uniqserv.NetworkLabels {
					if anotherNetLabel == uniqservsNetLabel {
						continue andNowForSomethingCompletelyDifferent
					}
				}
				uniqserv.NetworkLabels = append(uniqserv.NetworkLabels, anotherNetLabel)
			}
			uniques[servicename] = uniqserv
		} else {
			uniques[servicename] = service
		}
	}
	uniqs := make(Services, 0, len(uniques)) // Did you mean "unix"? *snicker*
	for _, service := range uniques {
		uniqs = append(uniqs, service)
	}
	return uniqs
}

// FilterNetworkLabels ensures that a ServiceOnNetworks' networks only list
// those network labels specified.
func (s *Services) FilterNetworkLabels(netlabels []string) {
	// MIND THE PASS BY VALUE RANGE... ARGHHHHH, AGAIN.
	for sidx := range *s {
		// use in-place filtering
		networkLabels := (*s)[sidx].NetworkLabels
		idx, last := 0, 0
	filterNextNetworkLabel:
		for idx < len(networkLabels) {
			networkLabel := networkLabels[idx]
			for _, allowedNetworkLabel := range netlabels {
				if networkLabel == allowedNetworkLabel {
					networkLabels[last] = networkLabels[idx]
					idx++
					last++
					continue filterNextNetworkLabel
				}
			}
			idx++
		}
		(*s)[sidx].NetworkLabels = networkLabels[:last]
	}
}

// DnsNames returns all DNS names for addressing this service from the point of
// a particular container.
func (s ServiceOnNetworks) DnsNames() []string {
	var names []string

	// Service names, if any, out of and in the contexts of network names.
	if s.Name != "" {
		names = append(names, s.Name)
		for _, netlabel := range s.NetworkLabels {
			names = append(names, s.Name+"."+netlabel)
		}
	}
	// Individual container names out of and in the context of network names.
	for _, cntr := range s.Servants {
		names = append(names, cntr.Name)
		for _, netlabel := range s.NetworkLabels {
			names = append(names, cntr.Name+"."+netlabel)
		}
	}

	return names
}
