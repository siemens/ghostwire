// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build matchers
// +build matchers

package util

import (
	"fmt"

	"github.com/siemens/ghostwire/v2/network"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

// ContainContainer succeeds if ACTUAL is a slice, array or map of
// *network.NetworkNamespaces and contains the container with the specified name
// or ID and is also of the specified type. Alternatively of a name/ID string, a
// GomegaMatcher can also be specified for matching the name or ID, such as
// ContainSubstring and MatchRegexp.
func ContainContainer(nameorid interface{}, typ string) types.GomegaMatcher {
	return ContainElement(WithTransform(
		func(actual *network.NetworkNamespace) []*model.Container {
			containers := make([]*model.Container, 0, len(actual.Tenants))
			for _, tenant := range actual.Tenants {
				if container := tenant.Process.Container; container != nil {
					containers = append(containers, container)
				}
			}
			return containers
		},
		ContainElement(SatisfyAll(
			HaveContainerNameID(nameorid),
			HaveContainerType(typ),
		))))
}

// HaveContainer succeeds if ACTUAL is either a model.Container or
// *model.Container with the specified name or ID and also of the specified
// type.  Alternatively of a name/ID string, a GomegaMatcher can also be
// specified for matching the name or ID, such as ContainSubstring and
// MatchRegexp.
func HaveContainer(nameid interface{}, typ string) types.GomegaMatcher {
	return SatisfyAll(HaveContainerNameID(nameid), HaveContainerType(typ))
}

// HaveContainerNameID succeeds if ACTUAL is either a model.Container or
// *model.Container with the specified name or ID. Alternatively of a name/ID
// string, a GomegaMatcher can also be specified for matching the name or ID,
// such as ContainSubstring and MatchRegexp.
func HaveContainerNameID(nameorid interface{}) types.GomegaMatcher {
	var nameoridMatcher types.GomegaMatcher
	switch nameorid := nameorid.(type) {
	case string:
		nameoridMatcher = Equal(nameorid)
	case types.GomegaMatcher:
		nameoridMatcher = nameorid
	default:
		panic("nameorid argument must be string or GomegaMatcher")
	}
	return SatisfyAny(
		WithTransform(func(actual interface{}) (string, error) {
			switch container := actual.(type) {
			case *model.Container:
				return container.ID, nil
			case model.Container:
				return container.ID, nil
			}
			return "", fmt.Errorf("HaveContainerNameID expects a model.Container or *model.Container, but got %T", actual)
		}, nameoridMatcher),
		WithTransform(func(actual *model.Container) string { return actual.Name }, Equal(nameorid)),
	)
}

// HaveContainerType succeeds if ACTUAL is either a model.Container or
// *model.Container with the specified type.
func HaveContainerType(typ string) types.GomegaMatcher {
	return WithTransform(func(actual interface{}) string {
		switch container := actual.(type) {
		case *model.Container:
			return container.Type
		case model.Container:
			return container.Type
		}
		panic("HaveContainerType expects a model.Container or *model.Container")
	}, Equal(typ))
}

func FromPod(pod interface{}) types.GomegaMatcher {
	var podNameMatcher types.GomegaMatcher
	switch pod := pod.(type) {
	case string:
		podNameMatcher = Equal(pod)
	case types.GomegaMatcher:
		podNameMatcher = pod
	default:
		panic("pod argument must be string or GomegaMatcher")
	}
	return WithTransform(func(actual interface{}) ([]*model.Group, error) {
		switch container := actual.(type) {
		case *model.Container:
			return container.Groups, nil
		case model.Container:
			return container.Groups, nil
		}
		return nil, fmt.Errorf("FromPod expects a model.Container or *model.Container, but got %T", actual)
	}, ContainElement(
		WithTransform(
			func(group *model.Group) string {
				return group.Name
			}, podNameMatcher)))
}
