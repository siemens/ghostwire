// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/nonstd/xslog"
	"golang.org/x/sys/unix"

	gostwire "github.com/siemens/ghostwire/v2"
)

// gostwireservice is the "root command" to be run after successfully parsing
// the CLI flags. We then here kick off the Gostwire service itself.
func gostwireservice(cmd *cobra.Command, _ []string) error {
	// initial cgroup hack around docker-compose not allowing for setting
	// "cgroupns: host" during deployment. This is not necessary for newer
	// docker compose v2 plugins that implement the "cgroup: host" service
	// setting.
	switchedCgroup, _ := cmd.PersistentFlags().GetBool("cgroupswitched")
	switchCgroup, _ := cmd.PersistentFlags().GetBool("initialcgroup")
	if !switchedCgroup && switchCgroup {
		// Let's check if we're in a non-initial cgroup, such as when running
		// inside a Docker container on a "pure" unified cgroups v2 hierarchy
		// and with container cgroup'ing enabled...
		initialcgroupns := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP)
		initialcgroupnsid, ierr := initialcgroupns.ID()
		currentcgroupnsid, cerr := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		if ierr != nil || cerr != nil {
			slog.Error("cannot determine initial and own cgroup namespaces, not switching cgroup namespace")
		} else if currentcgroupnsid != initialcgroupnsid {
			// In order to safely switch the cgroup namespace in a Golang app
			// with potentially several OS threads bouncing around by now we can
			// only lock our current OS thread, switch into the initial cgroup
			// namespace, and finally reexecute ourselves again. We don't use
			// our re-execution support package here, as that on purpose is
			// designed to not be re-executable from a child and provides a
			// JSON-oriented result interface (which we don't need). Instead, we
			// just use the few and simple primitives, namely
			// runtime.LockOSThread() and ops.Execute(). Everything else is
			// debug logging and error handling.
			slog.Info("switching into initial cgroup and re-executing...",
				slog.Uint64("current.cgroup", currentcgroupnsid.Ino),
				slog.Uint64("initial.cgroup", initialcgroupnsid.Ino))
			runtime.LockOSThread()
			if res, err := ops.Execute(func() error {
				// tee hee, while the process might still be in its original,
				// but not initial, cgroup namespace, this particular OS-level
				// task/thread should now be in the initial cgroup namespace. So
				// we need to query the current cgroup namespace by TID, not
				// PID. Please note that we don't need to use
				// "/proc/self/task/$TID/ns/cgroup", as all tasks are directly
				// accessible at the /proc/$TID level; they're just not listed,
				// yet still there.
				currentcgroupnsid, _ := ops.NewTypedNamespacePath(
					fmt.Sprintf("/proc/%d/ns/cgroup", syscall.Gettid()),
					species.CLONE_NEWCGROUP).ID()
				if currentcgroupnsid == initialcgroupnsid {
					slog.Debug("current OS thread successfully switched cgroup namespace",
						slog.Uint64("cgroup", currentcgroupnsid.Ino))
				} else {
					slog.Debug("current OS thread cgroup namespace",
						slog.Uint64("cgroup", currentcgroupnsid.Ino))
				}
				return unix.Exec(
					"/proc/self/exe",
					append([]string{os.Args[0], "--cgroupswitched"}, os.Args[1:]...),
					os.Environ(),
				)
			}, initialcgroupns); err != nil {
				slog.Error("failed to switch to initial cgroup", xslog.Error(err))
			} else {
				slog.Error("failed to re-execute", xslog.Error(res))
				os.Exit(1)
			}
		}
	} else if switchedCgroup {
		// After re-execution log information about the current cgroup
		// namespace, which hopefully will be the initial cgroup namespace...
		initialcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		currentcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		slog.Info("re-executed",
			slog.Bool("initial", currentcgroupnsid == initialcgroupnsid),
			slog.Uint64("cgroup", currentcgroupnsid.Ino))
		// Unfortunately, we end up here with /proc/self/stat stating our
		// process name as "exe", because we executed our own executable. This
		// is not terribly useful and user/admin friendly, so we try to set our
		// own process name from our first command line argument.
		runtime.LockOSThread() // this still runs on the main thread...!
		proc := model.NewProcess(model.PIDType(os.Getpid()), false)
		procname := append([]byte(proc.Basename()), 0)
		ptr := unsafe.Pointer(&procname[0]) // #nosec G103
		// prctl(PR_SET_NAME, ...) will silently truncate any process name
		// deemed too long, see also:
		// https://man7.org/linux/man-pages/man2/prctl.2.html
		if _, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
			slog.Error("cannot fix process name", xslog.Error(syscall.Errno(errno)))
		} else {
			slog.Debug("fixed re-executed process name", slog.String("name", proc.Basename()))
		}
		runtime.UnlockOSThread()
	}

	// And now for the real meat.
	slog.Info("Gostwire \"The Sequel\" virtual network topology and configuration discovery service",
		slog.String("version", gostwire.SemVersion))
	slog.Info("Copyright (c) Siemens AG 2018-2026")

	// De-base64 the brand icon in case it had been base-64 encoded. The web UI
	// expects the brand icon to be SVG that then gets used in some places in
	// the UI.
	if *brandIcon != "" {
		if decodedIcon, err := base64.StdEncoding.DecodeString(*brandIcon); err == nil {
			*brandIcon = string(decodedIcon)
		}
	}

	if pausebin := mountineer.StandaloneSandboxBinary(); pausebin != "" {
		slog.Info("using optimized pandora's sandbox binary", slog.String("path", pausebin))
	}

	slog.Debug("using container engine \"Turtles Anywhere\" technology")
	turtlesctx, turtlescancel := context.WithCancel(context.Background())
	defer turtlescancel()
	cizer := turtles.Containerizer(turtlesctx, cmd)

	// prime the list of discovered engines in the background...
	slog.Debug("priming list of discovered container engines in background")
	go func() {
		_ = gostwire.Discover(turtlesctx, cizer, nil)
	}()

	// Fire up the service
	addr, _ := cmd.PersistentFlags().GetString("http")
	if _, err := startServer(addr, cmd, cizer); err != nil {
		slog.Error("cannot start service", xslog.Error(err))
		os.Exit(1)
	}
	stopit := make(chan os.Signal, 1)
	signal.Notify(stopit, syscall.SIGINT)
	signal.Notify(stopit, syscall.SIGTERM)
	signal.Notify(stopit, syscall.SIGQUIT)
	<-stopit
	maxwait, _ := cmd.PersistentFlags().GetDuration("shutdown")
	stopServer(maxwait)
	return nil
}

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "gostwire",
		Short:   "gostwire virtual network topology and configuration discovery service",
		Version: gostwire.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: gostwireservice,
	}

	// Sets up the flags.
	pf := rootCmd.PersistentFlags()
	pf.String("http", "[::]:5000", "HTTP service address")
	pf.Duration("shutdown", 15*time.Second, "graceful shutdown duration limit")

	// Work around docker-compose currently having no means to set "cgroupns:
	// host" during deployment. There's a CLI flag, but no docker-composer
	// support, see also docker/compose issue #8167:
	// https://github.com/docker/compose/issues/8167.
	pf.Bool("initialcgroup", false, "switches into initial cgroup namespace")
	pf.Bool("cgroupswitched", false, "")
	_ = pf.MarkHidden("cgroupswitched")
	// G(h)ostwire-specific CLI flags
	brandName = pf.StringP("brand", "", "Ghostwire", "brand name to show in the UI")
	brandIcon = pf.StringP("brandicon", "", "", "brand icon SVG markup (optionally base64 encoded)")

	clippy.AddFlags(rootCmd)

	return
}

var brandName *string
var brandIcon *string
