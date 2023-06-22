// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package metadata

import (
	"encoding/json"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testMetadataPluginName = "testmetadataneer"

type testMetadata1 struct {
	Bar string `json:"bar"`
}

type testMetadata2 struct {
	Baz string `json:"baz"`
}

func init() {
	plugger.Group[Metadata]().Register(func(gostwire.DiscoveryResult) map[string]interface{} {
		return map[string]interface{}{
			"testmeta": testMetadata1{Bar: "BAR"},
		}
	}, plugger.WithPlugin(testMetadataPluginName+"-2"))
	plugger.Group[Metadata]().Register(func(gostwire.DiscoveryResult) map[string]interface{} {
		return map[string]interface{}{
			"testmeta": testMetadata2{Baz: "BAZZ"},
		}
	}, plugger.WithPlugin(testMetadataPluginName+"-1"))
}

var _ = Describe("metadata", func() {

	It("turns structs into maps", func() {
		m, err := toMap(testMetadata1{Bar: "BAR"})
		Expect(err).NotTo(HaveOccurred())
		Expect(m).To(Equal(map[string]interface{}{
			"bar": "BAR",
		}))
	})

	It("merges maps deeply", func() {
		a := map[string]interface{}{
			"Foo": "foo",
			"Bar": map[string]interface{}{
				"Baz": "baz",
			},
		}
		b := map[string]interface{}{}
		deepMapMerge(a, b)
		Expect(b).To(Equal(a))
		a2 := map[string]interface{}{
			"Bar": map[string]interface{}{
				"Raz": "RAZ",
			},
		}
		deepMapMerge(a2, b)
		Expect(b).To(Equal(map[string]interface{}{
			"Foo": "foo",
			"Bar": map[string]interface{}{
				"Baz": "baz",
				"Raz": "RAZ",
			},
		}))
	})

	It("augments", func() {
		base := struct {
			Foo string `json:"foo"`
		}{Foo: "foo"}
		md, err := Augment(gostwire.DiscoveryResult{}, base)
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Marshal(md)).To(MatchJSON(`
{
	"foo": "foo",
	"testmeta": {
		"bar": "BAR",
		"baz": "BAZZ"
	}
}`))
	})

})
