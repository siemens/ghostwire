// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

import (
	"log/slog"
	"os"
	"path"
	"reflect"
	"time"

	lxkns "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/xsk"
	"github.com/siemens/ghostwire/v2/xsk/umem"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var packageTestName = func() string {
	type dummy struct{}
	_, name := path.Split(reflect.TypeFor[dummy]().PkgPath())
	name += ".test"
	if len(name) > 15 {
		return name[:15]
	}
	return name
}()

var _ = Describe("discovering umems", func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
			ShouldNot(HaveLeaked(goodgos))
		Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
	})

	It("collects the umems", func() {

		const (
			chunkAmount1 = 128
			chunkAmount2 = 256
			chunkSize    = 2048
		)

		if os.Getuid() != 0 {
			Skip("needs root")
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		By("entering a temporary network namespace and creating two umems and three XSKs")
		defer netns.EnterTransient()()

		dmy := dummy.NewTransientUp()
		mcvlan1 := macvlan.NewTransient(dmy)
		mcvlan2 := macvlan.NewTransient(dmy)

		umem1fd := Successful(umem.New(int64(chunkAmount1) * int64(chunkSize)))
		defer unix.Close(umem1fd)
		umem2fd := Successful(umem.New(int64(chunkAmount2) * int64(chunkSize)))
		defer unix.Close(umem2fd)

		xsk1a := Successful(xsk.New(
			mcvlan1.Attrs().Index,
			0,
			xsk.WithUmemFd(umem1fd),
			xsk.WithChunkAmount(chunkAmount1),
			xsk.WithChunkSize(chunkSize),
			xsk.WithFillRingSize(chunkAmount1),
			xsk.WithCompletionRingSize(chunkAmount1),
			xsk.WithoutRxRing(),
			xsk.WithTxRingSize(chunkAmount1),
		))
		defer xsk1a.Close()

		xsk1b := Successful(xsk.New(
			mcvlan1.Attrs().Index,
			0,
			xsk.WithSharedUmem(xsk1a),
			xsk.WithRxRingSize(chunkAmount1),
			xsk.WithoutTxRing(),
		))
		defer xsk1b.Close()

		xsk2 := Successful(xsk.New(
			mcvlan2.Attrs().Index,
			0,
			xsk.WithUmemFd(umem2fd),
			xsk.WithChunkAmount(chunkAmount2),
			xsk.WithChunkSize(chunkSize),
			xsk.WithFillRingSize(chunkAmount2),
			xsk.WithCompletionRingSize(chunkAmount2),
			xsk.WithRxRingSize(chunkAmount2),
			xsk.WithTxRingSize(chunkAmount2),
		))
		defer xsk2.Close()

		By("discovering our XSKs related to our process, with their umems")
		disco := lxkns.Namespaces(
			lxkns.WithStandardDiscovery(),
			lxkns.FromTasks(),
		)
		Expect(len(disco.Namespaces[model.NetNS])).To(BeNumerically(">=", 2))

		xsks := AllXSKs(disco)
		umems := ExtractUmems(xsks)
		Expect(len(umems)).To(BeNumerically(">=", 2))

		var umem1 *Umem
		Expect(umems).To(ContainElement(
			HaveField("FillRingEntries", uint32(xsk1a.Options().FillRingSize)),
			&umem1))
		Expect(umem1.Size).To(Equal(uint64(chunkAmount1) * uint64(chunkSize)))
		Expect(umem1.XSKs).To(HaveLen(2))
		Expect(umem1.XSKs).To(HaveEach(
			HaveField("Processes", ConsistOf(
				HaveField("Name", packageTestName)))))

		var umem2 *Umem
		Expect(umems).To(ContainElement(
			HaveField("FillRingEntries", uint32(xsk2.Options().FillRingSize)),
			&umem2))
		Expect(umem2.XSKs).To(HaveLen(1))
		Expect(umem2.XSKs).To(HaveEach(
			HaveField("Processes", ConsistOf(
				HaveField("Name", packageTestName)))))
	})

})
