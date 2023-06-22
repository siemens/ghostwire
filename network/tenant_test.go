// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

const testTenantWorkloadName = "gostwire-test-tenant-workload"

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

	It("discovers tenant's DNS configuration", NodeTimeout(30*time.Second), func(_ context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a test workload with specific DNS configuration")
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		testwl, err := pool.RunWithOptions(&dockertest.RunOptions{
			Privileged: true,
			Repository: "busybox",
			Tag:        "latest",
			Name:       testTenantWorkloadName,
			Cmd: []string{
				"/bin/sh", "-c",
				`
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
/bin/sleep 120s`,
			},
			Hostname: "foobar",
		})
		Expect(err).NotTo(HaveOccurred(), "container %s", testVethWorkloadName)
		defer testwl.Close()

		By("running a discovery")
		allnetns, lxknsdisco := discoverRedux()
		Expect(allnetns).NotTo(BeEmpty())

		netnsid := lxknsdisco.Processes[model.PIDType(testwl.Container.State.Pid)].Namespaces[model.NetNS].ID()
		netns := allnetns[netnsid]
		Expect(netns).NotTo(BeNil())
		Expect(netns.Tenants).To(HaveLen(1))

		tenant := netns.Tenants[0]
		Expect(tenant.DNS.Hostname).To(Equal("foobar"))
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
