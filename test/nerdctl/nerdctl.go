// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctl

import (
	"context"
	"os"
	"os/exec"
	"time"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	bareContainerdAPI   = "/run/containerd/containerd.sock"
	dockerContainerdAPI = "/var/run/docker/containerd/containerd.sock"
)

func containerdAPIPath() string {
	if os.Getuid() != 0 {
		return bareContainerdAPI
	}
	info, err := os.Stat(dockerContainerdAPI)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return bareContainerdAPI
	}
	return dockerContainerdAPI
}

// Nerdctl runs a nerdctl command with the specified CLI arguments, expecting
// the command to succeed without any error code.
func Nerdctl(ctx context.Context, args ...string) {
	gi.GinkgoHelper()
	args = append([]string{"--address", containerdAPIPath()}, args...)
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
	args = append([]string{"--address", containerdAPIPath()}, args...)
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
