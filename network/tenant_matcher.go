// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build matchers
// +build matchers

package network

import (
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

// ContainTenantWithPID succeeds if ACTUAL is of type network.Tenants and
// contains a tenant with the specified PID.
func ContainTenantWithPID(pid model.PIDType) types.GomegaMatcher {
	return g.WithTransform(
		func(actual Tenant) model.PIDType { return actual.Process.PID },
		g.ContainElement(pid))
}
