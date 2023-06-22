// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	_ "github.com/thediveo/lxkns/log/logrus"
)

func main() {
	// This is cobra boilerplate documentation, except for the missing call to
	// fmt.Println(err) which in the original boilerplate is just plain wrong:
	// it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
