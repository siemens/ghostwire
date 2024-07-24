// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cpus

import (
	"os"
	"strconv"
	"strings"

	"github.com/siemens/ghostwire/v2/metadata"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"

	gostwire "github.com/siemens/ghostwire/v2"
)

func init() {
	plugger.Group[metadata.Metadata]().Register(
		Metadata, plugger.WithPlugin("cpus"))
}

// Metadata returns metadata describing certain aspects of the host the
// discovery was run on, such as its host name, OS version, ...
func Metadata(r gostwire.DiscoveryResult) map[string]interface{} {
	onlinecpus, err := os.ReadFile("/sys/devices/system/cpu/online")
	if err != nil {
		log.Errorf("cannot retrieve list of online cpus, reason: %s", err.Error())
		return nil
	}
	cpulist := parseCPUList(strings.TrimSuffix(string(onlinecpus), "\n"))
	if cpulist == nil {
		log.Errorf("malformed /sys/devices/system/cpu/online")
		return nil
	}
	return map[string]interface{}{
		"cpus": cpulist,
	}
}

// parseCPUList parses the textual representation of a CPU list in form of
// comma-separated CPU ranges, returning the CPUList if successful; otherwise,
// it returns nil.
func parseCPUList(cpus string) (list model.CPUList) {
	for cpus != "" {
		cpurange := cpus
		sepIdx := strings.Index(cpus, ",")
		if sepIdx < 0 {
			// final range, so there won't be any further ranges
			cpus = ""
		} else {
			// intermediate range...
			cpurange = cpus[:sepIdx]
			cpus = cpus[sepIdx+1:]
			// ...there needs to be something following, so no hanging commas.
			if cpus == "" {
				return nil
			}
		}
		dashIdx := strings.Index(cpurange, "-")
		if dashIdx < 0 {
			// single CPU number
			cpuno, err := strconv.ParseUint(cpurange, 10, 32)
			if err != nil {
				return nil
			}
			list = append(list, [2]uint{uint(cpuno), uint(cpuno)})
			continue
		}
		// CPU number range
		fromcpu, err := strconv.ParseUint(cpurange[:dashIdx], 10, 32)
		if err != nil {
			return nil
		}
		tocpu, err := strconv.ParseUint(cpurange[dashIdx+1:], 10, 32)
		if err != nil {
			return nil
		}
		list = append(list, [2]uint{uint(fromcpu), uint(tocpu)})
	}
	return list
}
