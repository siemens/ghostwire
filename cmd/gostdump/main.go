// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"

	gostwire "github.com/siemens/ghostwire/v2"
	apiv1 "github.com/siemens/ghostwire/v2/api/v1"
	"github.com/siemens/turtlefinder"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cizer := turtlefinder.New(func() context.Context { return ctx })
	defer cancel()
	defer cizer.Close()

	discovery := gostwire.Discover(ctx, cizer, nil)

	result := apiv1.NewDiscoveryResult(discovery)
	j, err := json.Marshal(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}
