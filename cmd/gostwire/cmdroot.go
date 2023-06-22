// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/siemens/ghostwire/v2/turtlefinder"

	gostwire "github.com/siemens/ghostwire/v2"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// gostwireservice is the "root command" to be run after successfully parsing
// the CLI flags. We then here kick off the Gostwire service itself.
func gostwireservice(cmd *cobra.Command, _ []string) error {
	if silent, _ := cmd.PersistentFlags().GetBool("silent"); silent {
		log.SetLevel(log.ErrorLevel)
	}
	if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("gostwire service debug logging enabled")
	}
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
			log.Errorf("cannot determine initial and own cgroup namespaces, not switching cgroup namespace")
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
			log.Infof("switching from current cgroup:[%d] into initial cgroup:[%d] and re-executing...",
				currentcgroupnsid.Ino, initialcgroupnsid.Ino)
			runtime.LockOSThread()
			if res, err := ops.Execute(func() interface{} {
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
				success := ""
				if currentcgroupnsid == initialcgroupnsid {
					success = "successfully "
				}
				log.Debugf("current OS thread %sswitched to cgroup:[%d]", success, currentcgroupnsid.Ino)
				return unix.Exec(
					"/proc/self/exe",
					append([]string{os.Args[0], "--cgroupswitched"}, os.Args[1:]...),
					os.Environ(),
				)
			}, initialcgroupns); err != nil {
				log.Errorf("failed to switch to initial cgroup, err: %s", err.Error())
			} else {
				log.Errorf("failed to re-execute, err: %s", res)
				os.Exit(1)
			}
		}
	} else if switchedCgroup {
		// After re-execution log information about the current cgroup
		// namespace, which hopefully will be the initial cgroup namespace...
		initialcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		currentcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		isInitial := ""
		if currentcgroupnsid == initialcgroupnsid {
			isInitial = "initial "
		}
		log.Infof("re-executed in %scgroup:[%d]", isInitial, currentcgroupnsid.Ino)
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
			log.Errorf("cannot fix process name: %s", syscall.Errno(errno).Error())
		} else {
			log.Debugf("fixed re-executed process name to %q", proc.Basename())
		}
		runtime.UnlockOSThread()
	}

	// And now for the real meat.
	log.Infof("Gostwire \"The Sequel\" virtual network topology and configuration discovery service, version %s",
		gostwire.SemVersion)
	log.Infof("Copyright (c) Siemens AG 2018-2022")

	// De-base64 the brand icon in case it had been base-64 encoded. The web UI
	// expects the brand icon to be SVG that then gets used in some places in
	// the UI.
	if *brandIcon != "" {
		if decodedIcon, err := base64.StdEncoding.DecodeString(*brandIcon); err == nil {
			*brandIcon = string(decodedIcon)
		}
	}

	if pausebin := mountineer.StandaloneSandboxBinary(); pausebin != "" {
		log.Infof("using optimized pandora's sandbox binary %s", pausebin)
	}

	enginectx, enginecancel := context.WithCancel(context.Background())
	log.Debugf("using container engine \"Turtles Anywhere\" technology")
	cizer := turtlefinder.New(func() context.Context { return enginectx })
	defer enginecancel()

	// prime the list of discovered engines in the background...
	log.Debugf("priming list of discovered container engines in background")
	go func() {
		_ = gostwire.Discover(enginectx, cizer, nil)
	}()

	// Fire up the service
	addr, _ := cmd.PersistentFlags().GetString("http")
	if _, err := startServer(addr, cizer); err != nil {
		log.Errorf("cannot start service, error: %s", err.Error())
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
		RunE:    gostwireservice,
	}

	// Sets up the flags.
	pf := rootCmd.PersistentFlags()
	pf.Bool("debug", false, "enables debugging output")
	pf.Bool("silent", false, "silences everything below the error level")
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

	return
}

var brandName *string
var brandIcon *string
