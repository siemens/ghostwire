// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/thediveo/lxkns/model"
)

// pidNamespaces represents a set of PID namespaces and implements JSON
// marshalling.
type pidNamespaces model.NamespaceMap

// MarshalJSON marshals a set of PID namespaces in form of a JSON array of
// objects detailing the individual PID namespaces.
func (p pidNamespaces) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('[')
	first := true
	for _, pidns := range p {
		// At the top level only the root PID namespace ... that's the one
		// without any parent.
		if pidns.(model.Hierarchy).Parent() != nil {
			continue
		}
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		jpidns := newPidNamespace(pidns)
		jb, err := json.Marshal(jpidns)
		if err != nil {
			return nil, err
		}
		b.Write(jb)
	}
	b.WriteRune(']')
	return b.Bytes(), nil
}

func pidnsID(pidns model.Namespace) string {
	return "pidns-" + strconv.FormatUint(pidns.ID().Ino, 10)
}

// pidNamespace contains the JSON marshallable PID namespace information. Please
// note that the "netns-idref" information on purpose is absent: it is
// unreliable due to a misunderstanding of the (non-) relationships between
// network and PID namespaces when it comes to containers. We don't need them
// anyway nowadays.
type pidNamespace struct {
	ID              string         `json:"id"`
	PIDNsID         uint64         `json:"pidnsid"`
	Children        []pidNamespace `json:"children"`
	ContainerIDRefs []string       `json:"container-idrefs"`
}

func newPidNamespace(pidns model.Namespace) pidNamespace {
	pidnsChildren := pidns.(model.Hierarchy).Children()
	children := make([]pidNamespace, 0, len(pidnsChildren))
	for _, pidnsChild := range pidnsChildren {
		children = append(children, newPidNamespace(pidnsChild.(model.Namespace)))
	}
	crefs := make([]string, 0, len(pidns.Leaders()))
	for _, proc := range pidns.Leaders() {
		if proc.Container == nil {
			continue
		}
		crefs = append(crefs, cntrID(proc))
	}
	return pidNamespace{
		ID:              pidnsID(pidns),
		PIDNsID:         pidns.ID().Ino,
		Children:        children,
		ContainerIDRefs: crefs,
	}
}
