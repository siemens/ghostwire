// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"strings"

	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/turtlefinder"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
)

type captureTargets []captureTarget

type captureTarget struct {
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	NetnsID   uint64        `json:"netns"`
	NifNames  []string      `json:"network-interfaces"`
	PID       model.PIDType `json:"pid"`
	Starttime uint64        `json:"starttime"`
	Prefix    string        `json:"prefix"`
}

func newCaptureTargets(dr gostwire.DiscoveryResult) captureTargets {
	targets := make([]captureTarget, 0, len(dr.Netns))
	pods := map[string]struct{}{}
	for _, netns := range dr.Netns {
		nifs := netns.NifList()
		nifNames := make([]string, 0, len(nifs))
		for _, nif := range nifs {
			// To avoid problems with dumpcap denying/aborting capture for
			// certain types of network interfaces when these are in DOWN state
			// (especially TAP/TUN), skip them.
			if nif.Nif().State == network.Down {
				continue
			}
			nifNames = append(nifNames, nif.Nif().Name)
		}
		// Is this just a bind-mounted network namespace without any processes
		// (tenants)?
		if len(netns.Tenants) == 0 {
			targets = append(targets, captureTarget{
				Name:     strings.Join(netns.Ref(), ":"),
				Type:     "bindmount",
				NetnsID:  netns.ID().Ino,
				NifNames: nifNames,
			})
			continue
		}
		for _, tenant := range netns.Tenants {
			if tenant.Process.PPID == 0 && tenant.Process.PID == 2 {
				// skip kthreadd(2) in order to not bedazzle users.
				continue
			}
			// Is this tenant a stand-alone process?
			container := tenant.Process.Container
			if container == nil {
				targets = append(targets, captureTarget{
					Name:      tenant.Name(),
					Type:      "proc",
					NetnsID:   netns.ID().Ino,
					NifNames:  nifNames,
					PID:       tenant.Process.PID,
					Starttime: tenant.Process.Starttime,
					Prefix:    "", // never prefixed: not a container.
				})
				continue
			}
			// So this now is a container ... but: is it part of a pod?
			groups := container.Groups
			for _, group := range groups {
				if group.Type != kuhbernetes.PodGroupType {
					continue
				}
				container = nil
				// It's a container of a pod: if this pod is new to us, we
				// then take the pod as a capture target and otherwise skip
				// all containers in this pod.
				if _, ok := pods[group.Name]; ok {
					break
				}
				pods[group.Name] = struct{}{}
				targets = append(targets, captureTarget{
					Name:      group.Name,
					Type:      "pod",
					NetnsID:   netns.ID().Ino,
					NifNames:  nifNames,
					PID:       tenant.Process.PID,
					Starttime: tenant.Process.Starttime,
					Prefix:    "", // never prefixed: not a continer.
				})
				break
			}
			if container == nil {
				continue
			}
			// It's a stand-alone container.
			targets = append(targets, captureTarget{
				Name:      container.Name,
				Type:      v1ContainerType(container.Type),
				NetnsID:   netns.ID().Ino,
				NifNames:  nifNames,
				PID:       tenant.Process.PID,
				Starttime: tenant.Process.Starttime,
				Prefix:    container.Labels[turtlefinder.GostwireContainerPrefixLabelName],
			})
		}
	}
	return targets
}
