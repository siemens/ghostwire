// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import "encoding/json"

// PlainNetdev aliases the Netdev type, but without inheriting its
// un/marshalling methods, and used in custom Netdev unmarshalling.
//
// See also: https://choly.ca/post/go-json-marshalling/
type PlainNetdev Netdev

// jsonNetdev replaces Queues and NAPIs with their counterparts that un/marshal
// NAPI and IRQ pointers as plain IDs instead.
type jsonNetdev struct {
	*PlainNetdev
	Queues []*jsonQueue       `json:"queues"`
	NAPIs  map[uint]*jsonNAPI `json:"napis"`
	// IRQs can be directly unmarshalled, as process reference is resolved
	// later.
}

// UnmarshalJSON unmarshals a Netdev, regenerating Queue and NAPI pointers from
// IDs.
func (nd *Netdev) UnmarshalJSON(b []byte) error {
	var jnetdev jsonNetdev
	if err := json.Unmarshal(b, &jnetdev); err != nil {
		return err
	}

	// Phase I: gather and index the queues and NAPIs; we can unmarshal IRQs
	// directly, as they lack Queue and NAPI pointers. The kernel thread Process
	// reference is later set outside unmarshalling.
	*nd = Netdev(*jnetdev.PlainNetdev)
	nd.Queues = make([]*Queue, len(jnetdev.Queues))
	for idx, jQueue := range jnetdev.Queues {
		nd.Queues[idx] = (*Queue)(jQueue.PlainQueue)
	}
	nd.NAPIs = map[uint]*NAPI{}
	for id, jNAPI := range jnetdev.NAPIs {
		nd.NAPIs[id] = (*NAPI)(jNAPI.PlainNAPI)
	}

	// Phase II: resolve relationships between queues, NAPIs, and IRQs...
	for idx, jQueue := range jnetdev.Queues {
		if jQueue.NAPI != nil {
			nd.Queues[idx].NAPI = nd.NAPIs[*jQueue.NAPI]
		}
		if jQueue.IRQ != nil {
			nd.Queues[idx].IRQ = nd.IRQs[*jQueue.IRQ]
		}
	}
	for id, jNAPI := range jnetdev.NAPIs {
		if jNAPI.IRQ != nil {
			nd.NAPIs[id].IRQ = nd.IRQs[*jNAPI.IRQ]
		}
	}

	return nil
}

// MarshalJSON marshals a NAPI object into JSON. It replaces IRQ object pointers
// with IRQ IDs/numbers instead, breaking cycles.
func (n *NAPI) MarshalJSON() ([]byte, error) {
	jNAPI := &jsonNAPI{
		PlainNAPI: (*PlainNAPI)(n),
	}
	if n.IRQ != nil {
		jNAPI.IRQ = &n.IRQ.ID
	}
	return json.Marshal(jNAPI)
}

// MarshalJSON marshals a Queue object into JSON. It replaces NAPI and IRQ
// object pointers with their respective IDs.
func (q *Queue) MarshalJSON() ([]byte, error) {
	jQueue := &jsonQueue{
		PlainQueue: (*PlainQueue)(q),
	}
	if q.NAPI != nil {
		jQueue.NAPI = &q.NAPI.ID
	}
	if q.IRQ != nil {
		jQueue.IRQ = &q.IRQ.ID
	}
	return json.Marshal(jQueue)
}

// PlainNAPI aliases the NAPI type, but without inheriting its un/marshalling
// methods.
//
// See also: https://choly.ca/post/go-json-marshalling/
type PlainNAPI NAPI

// jsonNAPI replaces the IRQ reference with its respective IRQ ID instead.
type jsonNAPI struct {
	*PlainNAPI
	IRQ *uint `json:"irq,omitempty"`
}

// PlainQueue aliases the Queue type, but without inheriting its un/marshalling
// methods.
//
// See also: https://choly.ca/post/go-json-marshalling/
type PlainQueue Queue

// jsonQueue replaces NAPI and IRQ references with their respective IDs instead.
type jsonQueue struct {
	*PlainQueue
	NAPI *uint `json:"napi-id,omitempty"` // replaces NAPI pointer with ID
	IRQ  *uint `json:"irq,omitempty"`     // replaces IRQ pointer with ID
}
