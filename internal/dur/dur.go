// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package dur

import (
	"time"

	"github.com/thediveo/lxkns/log"
)

// SecMs return the seconds and milliseconds remainder for a given duration.
func SecMs(d time.Duration) (s, ms uint) {
	s = uint(d / time.Second)
	ms = uint((d % time.Second) / time.Millisecond)
	return
}

// Log durations at the information level.
func Log(what string, d time.Duration) {
	s, ms := SecMs(d)
	log.Infof("%s in %ds%03dms", what, s, ms)
}
