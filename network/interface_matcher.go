// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build matchers
// +build matchers

package network

import (
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// HaveInterfaceOfKindWithName succeeds if ACTUAL is of type network.Interface
// and is of the specified kind, as well as has the specified name.
func HaveInterfaceOfKindWithName(kind string, name string) types.GomegaMatcher {
	return g.SatisfyAll(
		HaveInterfaceName(name),
		g.WithTransform(
			func(actual Interface) string { return actual.Nif().Kind },
			g.Equal(kind)),
	)
}

// HaveInterfaceName succeeds if ACTUAL is of type network.Interface and has the
// specified name.
func HaveInterfaceName(name string) types.GomegaMatcher {
	return g.WithTransform(
		func(actual Interface) string { return actual.Nif().Name },
		g.Equal(name))
}

// ContainInterfaceWithName succeeds if ACTUAL is an array or map of
// network.Interfaces and contains a network interface with the specified name.
func ContainInterfaceWithName(name string) types.GomegaMatcher {
	return g.ContainElement(
		g.WithTransform(
			func(actual Interface) string { return actual.Nif().Name },
			g.Equal(name)))
}

// HaveInterfaceAlias succeeds if ACTUAL is of type network.Interface and has
// the specified alias (name).
func HaveInterfaceAlias(alias string) types.GomegaMatcher {
	return g.WithTransform(
		func(actual Interface) string { return actual.Nif().Alias },
		g.Equal(alias))
}

// ContainInterfaceWithAlias succeeds if ACTUAL is an array or map of
// network.Interface and contains a network interface with the specified alias
// name.
func ContainInterfaceWithAlias(name string) types.GomegaMatcher {
	return g.ContainElement(
		g.WithTransform(
			func(actual Interface) string { return actual.Nif().Alias },
			g.Equal(name)))
}
