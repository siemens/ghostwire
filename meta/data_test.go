// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package meta

import (
	"encoding/json"

	"github.com/thediveo/go-plugger/v3"

	gostwire "github.com/siemens/ghostwire/v2"

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

var _ = Describe("metadata", Ordered, func() {

	BeforeAll(func() {
		prods := plugger.Group[Producer]()
		DeferCleanup(prods.Restore, prods.Backup())
		prods.Clear()

		plugger.Group[Producer]().Register(func(gostwire.DiscoveryResult) Data {
			return Data{
				"testmeta": testMetadata1{Bar: "BAR"},
			}
		}, plugger.WithPlugin(testMetadataPluginName+"-2"))
		plugger.Group[Producer]().Register(func(gostwire.DiscoveryResult) Data {
			return Data{
				"testmeta": testMetadata2{Baz: "BAZZ"},
			}
		}, plugger.WithPlugin(testMetadataPluginName+"-1"))
	})

	It("turns structs into maps", func() {
		m, err := asData(testMetadata1{Bar: "BAR"})
		Expect(err).NotTo(HaveOccurred())
		Expect(m).To(Equal(Data{
			"bar": "BAR",
		}))
	})

	It("merges maps deeply", func() {
		a := Data{
			"Foo": "foo",
			"Bar": Data{
				"Baz": "baz",
			},
		}
		b := Data{}
		deepMerge(a, b)
		Expect(b).To(Equal(a))
		a2 := Data{
			"Bar": Data{
				"Raz": "RAZ",
			},
		}
		deepMerge(a2, b)
		Expect(b).To(Equal(Data{
			"Foo": "foo",
			"Bar": Data{
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
