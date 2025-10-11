// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"context"
	"os"
	"path"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/netns"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("network namespace-tagged sysfs", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	It("rejects a nil network namespace", func() {
		Expect(SysfsOfNetns(nil)).To(BeNil())
	})

	It("rejects a process-less network namespace", func() {
		emptynetnsfd := netns.NewTransient()
		defer unix.Close(emptynetnsfd)

		allns := discover.Namespaces(discover.WithStandardDiscovery())
		emptynetns := allns.Namespaces[model.NetNS][species.NamespaceIDfromInode(netns.Ino(emptynetnsfd))]
		Expect(emptynetns).NotTo(BeNil())
		Expect(SysfsOfNetns(emptynetns)).To(BeNil())
	})

	// Admittedly, this is a slightly hardcore test...
	It("rejects a network namespace without a matching mount namespace", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()

		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -n $stage2
`)
		scripts.Script("stage2", `
process_namespaceid net # prints the "current" net namespace ID.
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		netnsid := nstest.CmdDecodeNSId(cmd)

		allns := discover.Namespaces(discover.WithStandardDiscovery())
		emptynetns := allns.Namespaces[model.NetNS][netnsid]
		Expect(emptynetns).NotTo(BeNil())
		Expect(SysfsOfNetns(emptynetns)).To(BeNil())
	})

	It("returns a suitable Mountineer with the correct view", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("morbyd.test=sysfsguess.sysfs")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		sleepy := Successful(sess.Run(ctx,
			"busybox",
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter))))
		sleepyPID := Successful(sleepy.PID(ctx))

		allns := discover.Namespaces(discover.WithStandardDiscovery())
		Expect(allns.Processes).To(HaveKey(model.PIDType(sleepyPID)))
		sleepyNetns := allns.Processes[model.PIDType(sleepyPID)].Namespaces[model.NetNS]
		Expect(sleepyNetns).NotTo(BeNil())

		sysfsMounty := SysfsOfNetns(sleepyNetns)
		Expect(sysfsMounty).NotTo(BeNil())
		defer sysfsMounty.Close()

		sleepyNetnsFd, closeNetnsFd := Successful2R(ops.NamespacePath(sleepyNetns.Ref()[0]).NsFd())
		defer closeNetnsFd()
		dmy := dummy.NewTransient(dummy.InNamespace(sleepyNetnsFd))

		sysClassNetDummy := path.Join("/sys/class/net", dmy.Attrs().Name)
		Expect(sysClassNetDummy).NotTo(BeAnExistingFile(), "canary netdev should not appear in host view")
		Expect(sysfsMounty.Resolve(sysClassNetDummy)).To(BeADirectory(), "canary netdev cannot be found in container view")
	})

})
