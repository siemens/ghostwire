// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build kind
// +build kind

package kind

import (
	"context"
	"fmt"
	"os"

	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/turtlefinder"
	"github.com/siemens/ghostwire/v2/util"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/moby"
	"sigs.k8s.io/kind/pkg/cluster"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

const twoClusterConfig = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
`

const kindTestClusterName = "gw-kind-test"

var _ = Describe("kind", func() {

	var prov *cluster.Provider

	var cizer containerizer.Containerizer
	var ctx context.Context
	var cancel context.CancelFunc

	BeforeEach(func() {
		cancel = nil
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		// Spin up a new and dedicated KinD test cluster, but only if necessary.
		prov = cluster.NewProvider(cluster.ProviderWithDocker())
		Expect(prov).NotTo(BeNil())
		kinds := Successful(prov.List())
		preloaded := Successful(ContainElement(kindTestClusterName).Match(kinds))
		if !preloaded {
			Expect(
				prov.Create(kindTestClusterName, cluster.CreateWithRawConfig([]byte(twoClusterConfig))),
			).To(Succeed())
			DeferCleanup(func() {
				Expect(prov.Delete(kindTestClusterName, "")).To(Succeed())
			})
		}

		// Run a discovery...
		watcherctx, watchercancel := context.WithCancel(context.Background())
		cizer = turtlefinder.New(func() context.Context { return watcherctx })
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			watchercancel()
			cancel()
		})
	})

	Context("with a KinD cluster", func() {

		It("discovers node containers", func() {
			k8sNodeNames := Successful(prov.ListNodes(kindTestClusterName))
			Expect(k8sNodeNames).NotTo(BeEmpty())

			Eventually(func() []*model.Container {
				_, disco := discover.Discover(ctx, cizer, nil)
				return disco.Containers
			}, "10s", "500ms").Should(ContainElements(
				SatisfyAll(
					util.HaveContainer(kindTestClusterName+"-control-plane", moby.Type),
					turtlefinder.WithPrefix("")),
				SatisfyAll(
					util.HaveContainer(kindTestClusterName+"-worker", moby.Type),
					turtlefinder.WithPrefix("")),
			))
		})

		It("discovers KinD pod workloads", func() {
			var containers []*model.Container
			Eventually(func() []*model.Container {
				_, disco := discover.Discover(ctx, cizer, nil)
				containers = disco.Containers
				return containers
			}, "10s", "500ms").Should(ContainElements(
				SatisfyAll(
					util.FromPod(MatchRegexp(`^kube-system/etcd-%s-control-plane$`, kindTestClusterName)),
					turtlefinder.WithPrefix(kindTestClusterName+"-control-plane")),
				SatisfyAll(
					util.FromPod(MatchRegexp(`^kube-system/kube-proxy-\w+`)),
					turtlefinder.WithPrefix(kindTestClusterName+"-control-plane")),
				SatisfyAll(
					util.FromPod(MatchRegexp(`^kube-system/coredns-\w+-\w+$`)),
					turtlefinder.WithPrefix(kindTestClusterName+"-control-plane")),
				SatisfyAll(
					util.FromPod(MatchRegexp(`^kube-system/kube-proxy-\w+`)),
					turtlefinder.WithPrefix(kindTestClusterName+"-worker")),
			), func() string {
				return fmt.Sprintf("current pod list: %v", util.AllPods(containers))
			})
		})

	})

})
