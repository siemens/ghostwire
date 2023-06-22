// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"runtime"

	_ "github.com/thediveo/lxkns/log/logrus"
)

func init() {
	// lock the initial OS thread (that is, the Linux task group leader
	// representing the whole process) "M0" to the initial goroutine "G0". This
	// avoids that M0 ever gets scheduled to one of the goroutines that call
	// ops.Exec in order to execute a function in a different network namespace
	// and afterwards throw away the thread M scheduled and locked to such a
	// goroutine.
	//
	// For the details and background information, please see:
	//
	// golang/go issue 53210: "runtime: on Linux, better do not treat the
	// initial thread/task group leader as any other thread/task",
	// https://github.com/golang/go/issues/53210
	//
	// Google Groups go-nuts discussion "LockOSThread, switching (Linux kernel)
	// namespaces: what happens to the main thread...?",
	// https://groups.google.com/g/golang-nuts/c/dx-jweSVxHk
	runtime.LockOSThread()
}

func main() {
	// This is cobra boilerplate documentation, except for the missing call to
	// fmt.Println(err) which in the original boilerplate is just plain wrong:
	// it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
