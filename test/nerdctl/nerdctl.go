// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctl

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

// Nerdctl runs a nerdctl command with the specified CLI arguments, expecting
// the command to succeed without any error code.
func Nerdctl(args ...string) {
	session, err := gexec.Start(
		exec.Command("nerdctl", args...),
		gi.GinkgoWriter,
		gi.GinkgoWriter)
	g.ExpectWithOffset(1, err).NotTo(g.HaveOccurred())
	g.EventuallyWithOffset(1, session, "5s").Should(gexec.Exit(0))
}

// NerdctlIgnore runs a nerdctl command with the specified CLI arguments and
// ignores whatever outcome of running the nerdctl command will be.
func NerdctlIgnore(args ...string) {
	session, err := gexec.Start(
		exec.Command("nerdctl", args...),
		gi.GinkgoWriter,
		gi.GinkgoWriter)
	if err != nil {
		return
	}
	g.EventuallyWithOffset(1, session, "5s").Should(gexec.Exit())
}

// SkipWithout skips a test if nerdctl cannot be found in PATH.
func SkipWithout() {
	if _, err := exec.LookPath("nerdctl"); err != nil {
		gi.Skip("needs nerdctl in PATH")
	}
}
