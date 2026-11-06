// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	"github.com/thediveo/enumflag/v2"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/netdev/nlnetdev"
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/siemens/ghostwire/v2/netdev/sysfsguess"
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

// ModeFlag is the name of the discovery mode CLI flag.
const ModeFlag = "mode"

// newRootCmd returns the Cobra "root" command, thus the lsnetdev tool itself.
func newRootCmd() (rootCmd *cobra.Command) {
	var mode Mode
	rootCmd = &cobra.Command{
		Use:     "lsnetdev",
		Short:   "list netdev queue and IRQ information",
		Version: gostwire.SemVersion,
		Args:    cobra.NoArgs,
		// See: https://github.com/spf13/cobra/issues/340
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return clippy.BeforeCommand(rootCmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return discoverNetdevs(cmd, mode)
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().Var(
		enumflag.New(&mode, ModeFlag, ModeEnumMapping, enumflag.EnumCaseInsensitive),
		ModeFlag, "discovery mode; can be 'sysfs'/'sf', 'netlink'/'nl', or 'both'/'b'")
	clippy.AddFlags(rootCmd) // ...runs all registered SetupCLI plugins.
	return
}

func discoverNetdevs(cmd *cobra.Command, mode Mode) error {
	cizer := turtles.Containerizer(cmd.Context(), cmd)
	defer cizer.Close()
	allns := discover.Namespaces(
		discover.WithStandardDiscovery(),
		discover.WithContainerizer(cizer),
		discover.WithPIDMapper(), // recommended when using WithContainerizer.
		discover.FromTasks(),
	)
	var netdevmap rxtxlayout.NetdevsByNetns
	var err error
	switch mode {
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
		return fmt.Errorf("unsupported discovery mode %v", mode)
	}
	if err != nil {
		return fmt.Errorf("cannot discover netdevs, reason: %w", err)
	}

	// Sort network namespaces by their inode numbers; a side effect is that the
	// initial network namespace comes first, as it gets one of the lowest nsfs
	// ino numbers assigned.
	netnsids := slices.Collect(maps.Values(allns.Namespaces[model.NetNS]))
	slices.SortFunc(netnsids,
		func(netnsA, netnsB model.Namespace) int {
			return int(netnsA.ID().Ino) - int(netnsB.ID().Ino)
		})
	out := cmd.OutOrStdout()
	first := true
	for _, netnsid := range netnsids {
		netns := allns.Namespaces[model.NetNS][netnsid.ID()]
		// skip this network namespace if we haven't discovered any netdevs for
		// some strange reason...
		if netdevs, ok := netdevmap[netns]; !ok || len(netdevs) == 0 {
			continue
		}

		if !first {
			_, _ = fmt.Fprintln(out)
		}
		first = false
		_, _ = fmt.Fprintf(out, "net:[%d]\n", netns.ID().Ino)

		_, _ = fmt.Fprintln(out, "  leader process(es)")
		leaders := slices.Clone(netns.Leaders())
		slices.SortFunc(leaders, func(lA, lB *model.Process) int {
			return int(lA.PID) - int(lB.PID)
		})
		for _, leader := range leaders {
			s := "    ⚙ "
			if container := leader.Container; container != nil {
				s += fmt.Sprintf("%s (%s)", container.Name, container.Flavor)
			}
			s += fmt.Sprintf(" %q (%d)", leader.Name, leader.PID)
			_, _ = fmt.Fprintln(out, s)
		}

		_, _ = fmt.Fprintln(out, "  netdev(s)")
		ndevs := slices.Clone(netdevmap[netns])
		slices.SortFunc(ndevs, rxtxlayout.CompareNetdevsByName)
		for _, ndev := range ndevs {
			_, _ = fmt.Fprintf(out, "    🕳 %q (ifindex %d) source %s\n",
				ndev.Name, ndev.Index, ndev.Source)

			// list only IRQs that aren't referenced by queues/NAPIs
			for _, irq := range ndev.IRQs {
				if slices.ContainsFunc(ndev.Queues,
					func(q *rxtxlayout.Queue) bool {
						return q.IRQ == irq
					}) {
					continue
				}
				_, _ = fmt.Fprintf(out, "        🗲 IRQ %d", irq.ID)
				if irq.Kthread != nil {
					_, _ = fmt.Fprintf(out, " 🧵 kthread %q (PID %d)",
						irq.Kthread.Name, irq.Kthread.PID)
					_ = irq.Kthread.RetrieveAffinity()
					if affinity := irq.Kthread.Affinity; affinity != nil {
						_, _ = fmt.Fprint(out, " CPU ♥️ "+affinity.String())
					}
				}
				_, _ = fmt.Fprintf(out, " source %s\n", irq.Source)
			}

			napis := slices.Collect(maps.Values(ndev.NAPIs))
			slices.SortFunc(napis,
				func(a, b *rxtxlayout.NAPI) int {
					return int(a.ID) - int(b.ID)
				})
			for _, napi := range napis {
				_, _ = fmt.Fprintf(out, "        NAPI %d", napi.ID)
				if napi.Kthread != nil {
					_, _ = fmt.Fprintf(out, " 🧵 kthread %q (PID %d)",
						napi.Kthread.Name, napi.Kthread.PID)
					_ = napi.Kthread.RetrieveAffinity()
					if affinity := napi.Kthread.Affinity; affinity != nil {
						_, _ = fmt.Fprint(out, " CPU ♥️ "+affinity.String())
					}
				}
				if irq := napi.IRQ; irq != nil {
					_, _ = fmt.Fprintf(out, " IRQ %d", irq.ID)
					if irq.Source != ndev.Source {
						_, _ = fmt.Fprintf(out, " (source %s)", irq.Source)
					}
					if irq.Kthread != nil {
						_, _ = fmt.Fprintf(out, " 🧵 kthread %q (PID %d)",
							irq.Kthread.Name, irq.Kthread.PID)
						_ = irq.Kthread.RetrieveAffinity()
						if affinity := irq.Kthread.Affinity; affinity != nil {
							_, _ = fmt.Fprint(out, " CPU ♥️ "+affinity.String())
						}
					}
				} else {
					_, _ = fmt.Fprint(out, " IRQ -")
				}
				_, _ = fmt.Fprint(out, "\n")
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
				_, _ = fmt.Fprintf(out, "        %s %s queue %d", qt, queue.Type.String(), queue.ID)
				if queue.IRQ != nil {
					_, _ = fmt.Fprintf(out, " 🗲 IRQ %d", queue.IRQ.ID)
					if queue.IRQ.Source != ndev.Source {
						_, _ = fmt.Fprintf(out, " (source %s)", queue.IRQ.Source)
					}
				} else {
					_, _ = fmt.Fprint(out, " IRQ -")
				}
				if queue.NAPI != nil && queue.NAPI.Kthread != nil {
					_, _ = fmt.Fprintf(out, " NAPI 🏃 kthread %q (PID %d)", queue.NAPI.Kthread.Name, queue.NAPI.Kthread.PID)
				} else {
					_, _ = fmt.Fprint(out, " NAPI -")
				}
				_, _ = fmt.Fprint(out, "\n")
			}
		}
	}

	return nil
}
