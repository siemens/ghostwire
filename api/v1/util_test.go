// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ohler55/ojg/jp"

	. "github.com/onsi/gomega"
)

func validate(openapispec *openapi3.T, schemaname string, jsondata []byte) error {
	schemaref, ok := openapispec.Components.Schemas[schemaname]
	if !ok {
		return fmt.Errorf("invalid schema reference %q", schemaname)
	}
	var jsonobj interface{}
	if err := json.Unmarshal(jsondata, &jsonobj); err != nil {
		return err
	}
	return schemaref.Value.VisitJSON(jsonobj)
}

func jsnp(obj interface{}, expr string) interface{} {
	xpr, err := jp.ParseString(expr)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "invalid JSONPATH expression: %s", expr)
	r := xpr.Get(obj)
	ExpectWithOffset(1, r).To(HaveLen(1))
	return r[0]
}

func jsnpsl(obj interface{}, expr string) []interface{} {
	xpr, err := jp.ParseString(expr)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "invalid JSONPATH expression: %s", expr)
	r := xpr.Get(obj)
	return r
}
