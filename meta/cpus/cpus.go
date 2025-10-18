// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cpus

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/thediveo/cpus"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/nonstd/xslog"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/meta"
)

func init() {
	plugger.Group[meta.Producer]().Register(
		Metadata, plugger.WithPlugin("cpus"))
}

// Metadata returns metadata describing certain aspects of the host the
// discovery was run on, such as its host name, OS version, ...
func Metadata(r gostwire.DiscoveryResult) meta.Data {
	onlinecpus, err := os.ReadFile("/sys/devices/system/cpu/online")
	if err != nil {
		slog.Error("cannot retrieve list of online cpus", xslog.Error(err))
		return nil
	}
	cpulist, err := cpus.NewList(bytes.TrimSuffix(onlinecpus, []byte("\n")))
	if err != nil {
		slog.Error("malformed /sys/devices/system/cpu/online,", xslog.Error(err))
		return nil
	}
	return meta.Data{
		"cpus": cpulist,
	}
}
