// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package initialcgroup

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/spacetest"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("initial cgroup flags", func() {

	It("registers as CLI plugins, running after the debug plugin", func() {
		Expect(plugger.Group[cliplugin.SetupCLI]().Plugins()).To(ContainElement("ghostwire/initialcgroup"))
		plugins := plugger.Group[cliplugin.BeforeCommand]().Plugins()
		Expect(plugins).To(ContainElements("clippy/debug", "ghostwire/initialcgroup"))
		Expect(slices.Index(plugins, "clippy/debug")).To(
			BeNumerically("<", slices.Index(plugins, "ghostwire/initialcgroup")))
	})

	It("hides the re-execution flag", func() {
		root := &cobra.Command{Use: "root"}
		setupCLI(root)
		Expect(root.PersistentFlags().Lookup(InitialCgroupFlag)).To(HaveField("Hidden", BeFalse()))
		Expect(root.PersistentFlags().Lookup(CgroupSwitchedFlag)).To(HaveField("Hidden", BeTrue()))
	})

	It("does nothing unless asked to", func() {
		Expect(runBeforeCommand()).To(Succeed())
	})

	It("tidies up after re-execution", func() {
		Expect(runBeforeCommand("--" + CgroupSwitchedFlag)).To(Succeed())
	})

	When("asked to switch into the initial cgroup namespace", func() {

		var reexecArgs []string

		BeforeEach(func() {
			DeferCleanup(slog.SetDefault, slog.Default())
			slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))

			DeferCleanup(func(initial, own string, exec func(string, []string, []string) error) {
				initialCgroupnsPath, ownCgroupnsPath, reexec = initial, own, exec
			}, initialCgroupnsPath, ownCgroupnsPath, reexec)
			reexecArgs = nil
			reexec = func(_ string, argv []string, _ []string) error {
				reexecArgs = argv
				return errors.New("canary")
			}
		})

		It("doesn't re-execute when already in the initial cgroup namespace", func() {
			initialCgroupnsPath = "/proc/self/ns/cgroup"
			Expect(runBeforeCommand("--" + InitialCgroupFlag)).To(Succeed())
			Expect(reexecArgs).To(BeNil())
		})

		It("doesn't re-execute when the cgroup namespaces are unknown", func() {
			initialCgroupnsPath = "/nonexisting"
			Expect(runBeforeCommand("--" + InitialCgroupFlag)).To(Succeed())
			Expect(reexecArgs).To(BeNil())
		})

		It("doesn't re-execute when switching fails", func() {
			// has an "ID" differing from our cgroup namespace, yet can't be entered.
			initialCgroupnsPath = "/dev/null"
			Expect(runBeforeCommand("--" + InitialCgroupFlag)).To(Succeed())
			Expect(reexecArgs).To(BeNil())
		})

		It("re-executes in the initial cgroup namespace", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			cgroupnsfd := spacetest.NewTransient(unix.CLONE_NEWCGROUP)
			initialCgroupnsPath = fmt.Sprintf("/proc/self/fd/%d", cgroupnsfd)
			initialcgroupnsid := Successful(ops.NamespacePath(initialCgroupnsPath).ID())

			var reexecCgroupnsid species.NamespaceID
			reexec = func(_ string, argv []string, _ []string) error {
				reexecArgs = argv
				reexecCgroupnsid, _ = ops.NamespacePath("/proc/thread-self/ns/cgroup").ID()
				return errors.New("canary")
			}

			Expect(runBeforeCommand("--" + InitialCgroupFlag)).To(
				MatchError(ContainSubstring("failed to re-execute: canary")))
			Expect(reexecArgs).To(Equal(
				append([]string{os.Args[0], "--" + CgroupSwitchedFlag}, os.Args[1:]...)))
			Expect(reexecCgroupnsid).To(Equal(initialcgroupnsid))
		})

	})

})
