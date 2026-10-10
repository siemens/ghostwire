// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"context"
	"os"
	"reflect"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/net"
	"github.com/thediveo/morbyd/v2/net/bridge"
	"github.com/thediveo/morbyd/v2/net/macvlan"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/nonstd/xslices"
	nobridge "github.com/thediveo/notwork/bridge"
	"github.com/thediveo/notwork/dummy"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("network interface names of Docker networks", func() {

	const testnetworkName = "dockernet-decorator-notwork"

	BeforeEach(func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	DescribeTable("*vlan",
		func(ctx context.Context, driver string, explicitParent bool) {
			sess := Successful(morbyd.NewSession(ctx,
				session.WithAutoCleaning("test.decorator.dockernet.xvlan=")))
			DeferCleanup(sess.Close)

			if explicitParent {
				mastername := dummy.NewTransient().Attrs().Name
				netw := Successful(sess.CreateNetwork(ctx, testnetworkName,
					net.WithDriver(driver),
					macvlan.WithParent(mastername), // works also for ipvlan
				))
				DeferCleanup(netw.Remove)
				Expect(xvlanNifName(&netw.Details.Network.Network)).To(Equal(mastername))
				return
			}

			netw := Successful(sess.CreateNetwork(ctx, testnetworkName,
				net.WithDriver(driver),
			))
			DeferCleanup(netw.Remove)
			Expect(xvlanNifName(&netw.Details.Network.Network)).To(
				Equal("dm-" + netw.Details.Network.ID[0:12]))
		},
		Entry(nil, "ipvlan", true),
		Entry(nil, "ipvlan", false),
	)

	DescribeTable("bridges of the ancients",
		func(ctx context.Context, explicitBridge bool) {
			sess := Successful(morbyd.NewSession(ctx,
				session.WithAutoCleaning("test.decorator.dockernet.bridge=")))
			DeferCleanup(sess.Close)

			if explicitBridge {
				brname := nobridge.NewTransient().Attrs().Name
				netw := Successful(sess.CreateNetwork(ctx, testnetworkName,
					net.WithDriver("bridge"),
					bridge.WithBridgeName(brname),
				))
				DeferCleanup(netw.Remove)
				Expect(bridgeNifName(&netw.Details.Network.Network)).To(Equal(brname))
				return
			}

			netw := Successful(sess.CreateNetwork(ctx, testnetworkName,
				net.WithDriver("bridge"),
			))
			DeferCleanup(netw.Remove)
			Expect(bridgeNifName(&netw.Details.Network.Network)).To(
				Equal("br-" + netw.Details.Network.ID[0:12]))
		},
		Entry(nil, true),
		Entry(nil, false),
	)

	It("maps network interface names to their docker networks", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.decorator.dockernet.nifnamemapping=")))
		DeferCleanup(sess.Close)

		br := Successful(sess.CreateNetwork(ctx, testnetworkName+"-bridge",
			net.WithDriver("bridge")))
		DeferCleanup(br.Remove)
		ipvl := Successful(sess.CreateNetwork(ctx, testnetworkName+"-ipvlan",
			net.WithDriver("ipvlan")))
		DeferCleanup(ipvl.Remove)
		mcvl := Successful(sess.CreateNetwork(ctx, testnetworkName+"-macvlan",
			net.WithDriver("macvlan")))
		DeferCleanup(mcvl.Remove)

		notworks := xslices.Map(
			Successful(sess.Client().NetworkList(ctx, client.NetworkListOptions{})).Items,
			func(s network.Summary) network.Network { return s.Network })
		m := mapNifNamesToDockerNetworks(notworks)
		Expect(m).To(ContainElement(Satisfy(func(actual *network.Network) bool {
			return reflect.DeepEqual(*actual, br.Details.Network.Network)
		})))
	})

})
