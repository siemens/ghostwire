// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package dockerproxy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGostwireDecoratorDockerproxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/decorator/dockerproxy package")
}
