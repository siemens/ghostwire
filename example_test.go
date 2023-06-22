// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package gostwire

import (
	"context"
	"fmt"

	"github.com/siemens/ghostwire/v2/turtlefinder"
)

func Example_discovery() {
	enginectx, enginecancel := context.WithCancel(context.Background())
	defer enginecancel()
	containerizer := turtlefinder.New(func() context.Context { return enginectx })
	defer containerizer.Close()
	allnetns := Discover(context.Background(), containerizer, nil)
	fmt.Printf("%d network stacks found\n", len(allnetns.Netns))
}
