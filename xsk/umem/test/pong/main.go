// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

/*
pong is “player 2” in the pass-the-umem-as-a-file-descriptor game. It is
automatically run by the unit tests in the umem package.
*/
package main

import (
	"fmt"
	"time"

	"github.com/siemens/ghostwire/v2/xsk/umem"
)

// Parent will pass in the umem fd after stdin, stdout and stderr, so it will be
// number 3.
const umemfd = 3

func main() {
	umem, err := umem.Map(umemfd)
	if err != nil {
		panic(fmt.Errorf("cannot map fd %d, reason: %w", umemfd, err))
	}

	// "PLAYER 2 READY"
	umem[0] = 0x42

	// ...and PLAY!
	for {
		umem[1] = umem[0] + 1
		time.Sleep(1 * time.Millisecond)
	}
}
