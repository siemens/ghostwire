// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"

	gostwire "github.com/siemens/ghostwire/v2"
	apiv1 "github.com/siemens/ghostwire/v2/api/v1"
)

type Mode int

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:   "gostdump",
		Short: "dumps (full) discovery results in JSON REST v1 format",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: dumpDiscovery,
	}

	clippy.AddFlags(rootCmd)

	return rootCmd
}

func dumpDiscovery(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()
	discovery := gostwire.Discover(ctx, cizer, nil)
	result := apiv1.NewDiscoveryResult(discovery)
	j, err := json.Marshal(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	return nil
}
