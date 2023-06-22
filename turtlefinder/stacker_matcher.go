// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build matchers
// +build matchers

package turtlefinder

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/gomega"
)

func WithPrefix(prefix string) types.GomegaMatcher {
	return WithTransform(func(actual interface{}) (string, error) {
		switch container := actual.(type) {
		case *model.Container:
			return container.Labels[GostwireContainerPrefixLabelName], nil
		case model.Container:
			return container.Labels[GostwireContainerPrefixLabelName], nil
		}
		return "", fmt.Errorf("WithPrefix expects a model.Container or *model.Container, but got %T", actual)
	}, Equal(prefix))
}
