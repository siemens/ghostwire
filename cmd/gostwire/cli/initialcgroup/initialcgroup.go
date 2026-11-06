// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

package initialcgroup

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	_ "github.com/thediveo/clippy/debug" // placement anchor
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/nonstd/xslog"
	"golang.org/x/sys/unix"
)

// Names of the CLI flags provided in this package.
const (
	InitialCgroupFlag  = "initialcgroup"
	CgroupSwitchedFlag = "cgroupswitched"
)

// Test seams for the cgroup namespace references and re-execution.
var (
	initialCgroupnsPath = "/proc/1/ns/cgroup"
	ownCgroupnsPath     = "/proc/self/ns/cgroup"
	reexec              = unix.Exec
)

func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		setupCLI, plugger.WithPlugin("ghostwire/initialcgroup"))
	// Run after the debug plugin has set up logging.
	plugger.Group[cliplugin.BeforeCommand]().Register(
		beforeCommand, plugger.WithPlugin("ghostwire/initialcgroup"),
		plugger.WithPlacement(">clippy/debug"))
}

// setupCLI adds the initial cgroup flags.
func setupCLI(cmd *cobra.Command) {
	// Work around docker-compose currently having no means to set "cgroupns:
	// host" during deployment. There's a CLI flag, but no docker-composer
	// support, see also docker/compose issue #8167:
	// https://github.com/docker/compose/issues/8167.
	pf := cmd.PersistentFlags()
	pf.Bool(InitialCgroupFlag, false, "switches into initial cgroup namespace")
	pf.Bool(CgroupSwitchedFlag, false, "")
	_ = pf.MarkHidden(CgroupSwitchedFlag)
}

// beforeCommand switches into the initial cgroup namespace and re-executes
// this process when asked to, or otherwise tidies up after re-execution.
func beforeCommand(cmd *cobra.Command) error {
	if switched, _ := cmd.Flags().GetBool(CgroupSwitchedFlag); switched {
		tidyUpAfterReexecution()
		return nil
	}
	if switchCgroup, _ := cmd.Flags().GetBool(InitialCgroupFlag); switchCgroup {
		return switchIntoInitialCgroup()
	}
	return nil
}

// switchIntoInitialCgroup checks if we're in a non-initial cgroup, such as when
// running inside a Docker container on a "pure" unified cgroups v2 hierarchy
// and with container cgroup'ing enabled. If so, it switches into the initial
// cgroup namespace and re-executes this process. It only returns when not
// re-executing.
func switchIntoInitialCgroup() error {
	initialcgroupns := ops.NewTypedNamespacePath(initialCgroupnsPath, species.CLONE_NEWCGROUP)
	initialcgroupnsid, ierr := initialcgroupns.ID()
	currentcgroupnsid, cerr := ops.NewTypedNamespacePath(ownCgroupnsPath, species.CLONE_NEWCGROUP).ID()
	if ierr != nil || cerr != nil {
		slog.Error("cannot determine initial and own cgroup namespaces, not switching cgroup namespace")
		return nil
	}
	if currentcgroupnsid == initialcgroupnsid {
		return nil
	}
	// In order to safely switch the cgroup namespace in a Golang app with
	// potentially several OS threads bouncing around by now we can only lock
	// our current OS thread, switch into the initial cgroup namespace, and
	// finally reexecute ourselves again. We don't use our re-execution support
	// package here, as that on purpose is designed to not be re-executable from
	// a child and provides a JSON-oriented result interface (which we don't
	// need). Instead, we just use the few and simple primitives, namely
	// runtime.LockOSThread() and ops.Execute(). Everything else is debug
	// logging and error handling.
	slog.Info("switching into initial cgroup and re-executing...",
		slog.Uint64("current.cgroup", currentcgroupnsid.Ino),
		slog.Uint64("initial.cgroup", initialcgroupnsid.Ino))
	runtime.LockOSThread()
	res, err := ops.Execute(func() error {
		// tee hee, while the process might still be in its original, but not
		// initial, cgroup namespace, this particular OS-level task/thread
		// should now be in the initial cgroup namespace. So we need to query
		// the current cgroup namespace by TID, not PID. Please note that we
		// don't need to use "/proc/self/task/$TID/ns/cgroup", as all tasks are
		// directly accessible at the /proc/$TID level; they're just not
		// listed, yet still there.
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
		return reexec(
			"/proc/self/exe",
			append([]string{os.Args[0], "--" + CgroupSwitchedFlag}, os.Args[1:]...),
			os.Environ(),
		)
	}, initialcgroupns)
	if err != nil {
		slog.Error("failed to switch to initial cgroup", xslog.Error(err))
		return nil
	}
	return fmt.Errorf("failed to re-execute: %w", res)
}

// tidyUpAfterReexecution logs information about the current cgroup namespace,
// which hopefully will be the initial cgroup namespace, and fixes the process
// name.
func tidyUpAfterReexecution() {
	initialcgroupnsid, _ := ops.NewTypedNamespacePath(initialCgroupnsPath, species.CLONE_NEWCGROUP).ID()
	currentcgroupnsid, _ := ops.NewTypedNamespacePath(ownCgroupnsPath, species.CLONE_NEWCGROUP).ID()
	slog.Info("re-executed",
		slog.Bool("initial", currentcgroupnsid == initialcgroupnsid),
		slog.Uint64("cgroup", currentcgroupnsid.Ino))
	// Unfortunately, we end up here with /proc/self/stat stating our process
	// name as "exe", because we executed our own executable. This is not
	// terribly useful and user/admin friendly, so we try to set our own process
	// name from our first command line argument.
	runtime.LockOSThread() // this still runs on the main thread...!
	defer runtime.UnlockOSThread()
	proc := model.NewProcess(model.PIDType(os.Getpid()), false)
	procname := append([]byte(proc.Basename()), 0)
	ptr := unsafe.Pointer(&procname[0]) // #nosec G103
	// prctl(PR_SET_NAME, ...) will silently truncate any process name deemed
	// too long, see also: https://man7.org/linux/man-pages/man2/prctl.2.html
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
		slog.Error("cannot fix process name", xslog.Error(syscall.Errno(errno)))
		return
	}
	slog.Debug("fixed re-executed process name", slog.String("name", proc.Basename()))
}
