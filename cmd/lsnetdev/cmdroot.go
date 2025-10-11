// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/siemens/ghostwire/v2/cmd/internal/cli"
	_ "github.com/siemens/ghostwire/v2/cmd/internal/debug"
	"github.com/siemens/ghostwire/v2/cmd/internal/turtles"
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/siemens/ghostwire/v2/netdev/sysfsguess"
	"github.com/siemens/ghostwire/v2/nlnetdev"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"golang.org/x/exp/maps"
)

type Mode int

const (
	SysfsMode = iota
	NetlinkMode
	CombinedMode
)

var ModeEnumMapping = map[Mode][]string{
	SysfsMode:    {"sysfs", "sf"},
	NetlinkMode:  {"netlink", "nl"},
	CombinedMode: {"both", "b"},
}

var DisoveryMode Mode

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:   "lsnetdev",
		Short: "list netdev queue and IRQ information",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return cli.BeforeCommand(cmd)
		}, RunE: discoverNetdevs,
	}

	// Sets up the flags.
	pf := rootCmd.PersistentFlags()
	pf.Var(enumflag.New(&DisoveryMode, "mode", ModeEnumMapping, enumflag.EnumCaseInsensitive),
		"mode", "discovery mode; can be 'sysfs'/'sf', 'netlink'/'nl', or 'both'/'b'")

	cli.AddFlags(rootCmd)

	return rootCmd
}

func discoverNetdevs(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()
	allns := discover.Namespaces(
		discover.WithStandardDiscovery(),
		discover.WithContainerizer(cizer),
		discover.WithPIDMapper(), // recommended when using WithContainerizer.
		discover.FromTasks(),
	)
	var netdevmap rxtxlayout.NetdevsByNetns
	var err error
	switch DisoveryMode {
	case SysfsMode:
		netdevmap, err = sysfsguess.Discover(allns)
	case NetlinkMode:
		netdevmap, err = nlnetdev.Discover(allns)
	case CombinedMode:
		var sysfsnetdevmap, netlinknetdevmap rxtxlayout.NetdevsByNetns
		sysfsnetdevmap, err = sysfsguess.Discover(allns)
		if err == nil {
			netlinknetdevmap, err = nlnetdev.Discover(allns)
		}
		if err == nil {
			netdevmap = rxtxlayout.Fillin(netlinknetdevmap, sysfsnetdevmap)
		}
	default:
		return fmt.Errorf("unsupported error mode %v", DisoveryMode)
	}
	if err != nil {
		return fmt.Errorf("cannot discover netdevs, reason: %w", err)
	}

	// Sort network namespaces by their inode numbers; a side effect is that the
	// initial network namespace comes first, as it gets one of the lowest nsfs
	// ino numbers assigned.
	netnsids := maps.Values(allns.Namespaces[model.NetNS])
	slices.SortFunc(netnsids,
		func(netnsA, netnsB model.Namespace) int {
			return int(netnsA.ID().Ino) - int(netnsB.ID().Ino)
		})
	first := true
	for _, netnsid := range netnsids {
		netns := allns.Namespaces[model.NetNS][netnsid.ID()]
		if !first {
			fmt.Println()
		}
		first = false
		fmt.Printf("net:[%d]\n", netns.ID().Ino)

		fmt.Println("  leader process(es)")
		leaders := slices.Clone(netns.Leaders())
		slices.SortFunc(leaders, func(lA, lB *model.Process) int {
			return int(lA.PID) - int(lB.PID)
		})
		for _, leader := range leaders {
			s := "    🏃 "
			if container := leader.Container; container != nil {
				s += fmt.Sprintf("%s (%s)", container.Name, container.Flavor)
			}
			s += fmt.Sprintf(" %q (%d)", leader.Name, leader.PID)
			fmt.Println(s)
		}

		fmt.Println("  netdev(s)")
		ndevs := slices.Clone(netdevmap[netns])
		slices.SortFunc(ndevs, rxtxlayout.SortNetdevsByName)
		for _, ndev := range ndevs {
			fmt.Printf("    🔌 %q (ifindex %d) source %s\n",
				ndev.Name, ndev.Index, ndev.Source)

			// list only IRQs that aren't referenced by queues/NAPIs
			for _, irq := range ndev.IRQs {
				if slices.ContainsFunc(ndev.Queues,
					func(q *rxtxlayout.Queue) bool {
						return q.IRQ == irq
					}) {
					continue
				}
				fmt.Printf("        🗲 IRQ %d", irq.ID)
				if irq.Kthread != nil {
					fmt.Printf(" 🧵 kthread %q (PID %d)",
						irq.Kthread.Name, irq.Kthread.PID)
					_ = irq.Kthread.RetrieveAffinityScheduling()
					if affinity := irq.Kthread.Affinity; affinity != nil {
						fmt.Printf(" CPU ♥️ " + affinity.String())
					}
				}
				fmt.Printf(" source %s\n", irq.Source)
			}

			napis := maps.Values(ndev.NAPIs)
			slices.SortFunc(napis,
				func(a, b *rxtxlayout.NAPI) int {
					return int(a.ID) - int(b.ID)
				})
			for _, napi := range napis {
				fmt.Printf("        NAPI %d", napi.ID)
				if napi.Kthread != nil {
					fmt.Printf(" 🧵 kthread %q (PID %d)",
						napi.Kthread.Name, napi.Kthread.PID)
					_ = napi.Kthread.RetrieveAffinityScheduling()
					if affinity := napi.Kthread.Affinity; affinity != nil {
						fmt.Printf(" CPU♥️ " + affinity.String())
					}
				}
				if irq := napi.IRQ; irq != nil {
					fmt.Printf(" IRQ %d", irq.ID)
					if irq.Source != ndev.Source {
						fmt.Printf(" (source %s)", irq.Source)
					}
					if irq.Kthread != nil {
						fmt.Printf(" 🧵 kthread %q (PID %d)",
							irq.Kthread.Name, irq.Kthread.PID)
						_ = irq.Kthread.RetrieveAffinityScheduling()
						if affinity := irq.Kthread.Affinity; affinity != nil {
							fmt.Printf(" CPU♥️ " + affinity.String())
						}
					}
				} else {
					fmt.Printf(" IRQ -")
				}
				fmt.Printf("\n")
			}

			queues := slices.Clone(ndev.Queues)
			slices.SortFunc(queues, func(qA, qB *rxtxlayout.Queue) int {
				if delta := int(qA.ID) - int(qB.ID); delta != 0 {
					return delta
				}
				if qA.Type == rxtxlayout.NETDEV_QUEUE_TYPE_RX {
					return -1
				}
				return 1
			})
			for _, queue := range queues {
				qt := "📥"
				if queue.Type.String() == "TX" {
					qt = "📤"
				}
				fmt.Printf("        %s %s queue %d", qt, queue.Type.String(), queue.ID)
				if queue.IRQ != nil {
					fmt.Printf(" 🗲 IRQ %d", queue.IRQ.ID)
					if queue.IRQ.Source != ndev.Source {
						fmt.Printf(" (source %s)", queue.IRQ.Source)
					}
				} else {
					fmt.Printf(" IRQ -")
				}
				if queue.NAPI != nil && queue.NAPI.Kthread != nil {
					fmt.Printf(" NAPI 🏃 kthread %q (PID %d)", queue.NAPI.Kthread.Name, queue.NAPI.Kthread.PID)
				} else {
					fmt.Printf(" NAPI -")
				}
				fmt.Printf("\n")
			}
		}
	}

	return nil
}
