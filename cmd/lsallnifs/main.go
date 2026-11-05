// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func init() {
	// Lock the initial OS thread "M0" to the initial goroutine "G0", so that M0
	// never gets thrown away after switching namespaces, see also:
	// https://github.com/golang/go/issues/53210
	runtime.LockOSThread()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Cobra already reports any error, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		stop()
		os.Exit(1)
	}
}
