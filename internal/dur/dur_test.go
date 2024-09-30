// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package dur

import (
	"fmt"
	"time"

	"github.com/thediveo/lxkns/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testLogger struct {
	s string
}

var _ log.Logger = (*testLogger)(nil)

func (l *testLogger) Log(level log.Level, msg string) {
	l.s += fmt.Sprintf("[%d] %s\n", level, msg)
}

func (l *testLogger) SetLevel(level log.Level) {}

func (l *testLogger) String() string { return l.s }

var _ = Describe("durations", func() {

	DescribeTable("returns durations as seconds and milliseconds remainders",
		func(d time.Duration, expS, expMs int) {
			s, ms := SecMs(d)
			Expect(s).To(Equal(uint(expS)))
			Expect(ms).To(Equal(uint(expMs)))
		},
		Entry("zero", time.Duration(0), 0, 0),
		Entry("only milliseconds, no microseconds", 123*time.Millisecond+456*time.Microsecond, 0, 123),
		Entry("s and ms", 666*time.Second+123*time.Millisecond+456*time.Microsecond, 666, 123),
	)

	It("logs durations informally", func() {
		l := &testLogger{}
		log.SetLogger(l)
		Log("foobar", 123*time.Second+45*time.Millisecond)
		Expect(l.String()).To(Equal("[4] foobar in 123s045ms\n"))
	})

})
