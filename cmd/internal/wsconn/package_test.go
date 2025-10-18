// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package wsconn

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWsconn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/internal/wsconn package")
}
