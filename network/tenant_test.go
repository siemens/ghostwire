// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("tenant", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	It("discovers tenant's DNS configuration", NodeTimeout(30*time.Second), func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test=ghostwire.network")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating a test workload with specific DNS configuration")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithCombinedOutput(GinkgoWriter),
			run.WithCapAdd("CAP_SYS_ADMIN"),
			run.WithCommand("/bin/sh", "-c",
				`set -e &&
umount /etc/hostname && echo "etchostname" > /etc/hostname &&
echo "etcdomainname" > /etc/domainname &&
umount /etc/hosts && echo "# comment
127.1.2.3 frotzelwotz wurzelprumpf" > /etc/hosts &&
umount /etc/resolv.conf && echo "; comment
# comment
nameserver 1.2.3.4
nameserver 1,2
nameserver 1::xyz
nameserver ::dead:beef
domain abracadabra
search foo.bar frotz.batz
" > /etc/resolv.conf &&
while true; do sleep 1; done`),
		))
		cntrpid := Successful(cntr.PID(ctx))

		By("running a discovery")
		allnetns, lxknsdisco := discoverRedux()
		Expect(allnetns).NotTo(BeEmpty())

		netnsid := lxknsdisco.Processes[model.PIDType(cntrpid)].Namespaces[model.NetNS].ID()
		netns := allnetns[netnsid]
		Expect(netns).NotTo(BeNil())
		Expect(netns.Tenants).To(HaveLen(1))

		tenant := netns.Tenants[0]
		Expect(tenant.DNS.Hostname).To(Equal(cntr.ID[:12]))
		Expect(tenant.DNS.EtcHostname).To(Equal("etchostname"))
		Expect(tenant.DNS.EtcDomainname).To(Equal("etcdomainname"))

		Expect(tenant.DNS.Hosts).To(HaveLen(2))
		Expect(tenant.DNS.Hosts).To(HaveKeyWithValue("frotzelwotz", net.ParseIP("127.1.2.3").To4()))
		Expect(tenant.DNS.Hosts).To(HaveKeyWithValue("wurzelprumpf", net.ParseIP("127.1.2.3").To4()))

		Expect(tenant.DNS.Nameservers).To(ConsistOf(
			net.ParseIP("1.2.3.4").To4(),
			net.ParseIP("::dead:beef"),
		))

		Expect(tenant.DNS.Searchlist).To(ConsistOf("foo.bar", "frotz.batz"))
	})

})
