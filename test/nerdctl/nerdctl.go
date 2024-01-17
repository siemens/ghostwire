// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctl

import (
	"context"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gexec"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

// Nerdctl runs a nerdctl command with the specified CLI arguments, expecting
// the command to succeed without any error code.
func Nerdctl(ctx context.Context, args ...string) {
	gi.GinkgoHelper()
	session, err := gexec.Start(
		exec.Command("nerdctl", args...),
		gi.GinkgoWriter,
		gi.GinkgoWriter)
	g.Expect(err).NotTo(g.HaveOccurred())
	g.Eventually(ctx, session).ProbeEvery(100 * time.Millisecond).
		Should(gexec.Exit(0))
}

// NerdctlIgnore runs a nerdctl command with the specified CLI arguments and
// ignores whatever outcome of running the nerdctl command will be.
func NerdctlIgnore(ctx context.Context, args ...string) {
	gi.GinkgoHelper()
	session, err := gexec.Start(
		exec.Command("nerdctl", args...),
		gi.GinkgoWriter,
		gi.GinkgoWriter)
	if err != nil {
		return
	}
	g.Eventually(ctx, session).Should(gexec.Exit())
}

// SkipWithout skips a test if nerdctl cannot be found in PATH.
func SkipWithout() {
	if _, err := exec.LookPath("nerdctl"); err != nil {
		gi.Skip("needs nerdctl in PATH")
	}
}
