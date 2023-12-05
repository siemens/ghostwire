// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/decorator/ieappicon"
	"github.com/siemens/ghostwire/v2/test/nerdctl"
	"github.com/siemens/ghostwire/v2/util"
	"github.com/siemens/turtlefinder"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/containerd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	podNetworkName = "gw-target-test-network"
	podNamespace   = "wuergspace"
	podName        = "gw-target-test"
	podFQDN        = podNamespace + "/" + podName
	podc1          = "incontinentainer-1"
	podc2          = "incontinentainer-2"
	bareName       = "gw-target-test-workload"
	barePrefix     = "livinginabox"
)

func TestGostwireApiV1(t *testing.T) {
	openapi3.DefineIPv4Format()
	openapi3.DefineIPv6Format()

	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/api/v1 package")
}

// Register testing pod grouping plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("kuhtainerd"))
}

// PodDecorate decorates the discovered containerd containers with pod groups,
// where applicable. This used to be in general use until the advent of the CRI
// API-based watcher and its generic decoration; now we use it to avoid having
// to rewrite the whole test here.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// If it "ain't no" containerd, skip it, as we're looking specifically
		// for containerd engines and their particular Kubernetes pod labelling.
		if engine.Type != containerd.Type {
			continue
		}
		// Pods cannot span container engines ;)
		podgroups := map[string]*model.Group{}
		for _, container := range engine.Containers {
			podNamespace := container.Labels[kuhbernetes.PodNamespaceLabel]
			podName := container.Labels[kuhbernetes.PodNameLabel]
			if podName == "" || podNamespace == "" {
				continue
			}
			// Create a new pod group, if it doesn't exist yet. Add the
			// container to its pod group.
			namespacedpodname := podNamespace + "/" + podName
			podgroup, ok := podgroups[namespacedpodname]
			if !ok {
				podgroup = &model.Group{
					Name:   namespacedpodname,
					Type:   kuhbernetes.PodGroupType,
					Flavor: kuhbernetes.PodGroupType,
				}
				podgroups[namespacedpodname] = podgroup
				total++
			}
			podgroup.AddContainer(container)
			// Sandbox? Then tag (label) the container.
			/*
				if container.Labels[containerKindLabel] == "sandbox" {
					container.Labels[kuhbernetes.PodSandboxLabel] = ""
				}
			*/
		}
	}
}

func tabulaRasa() {
	nerdctl.NerdctlIgnore("rm", "-f", podc1)
	nerdctl.NerdctlIgnore("rm", "-f", podc2)
	nerdctl.NerdctlIgnore("rm", "-f", bareName)
	nerdctl.NerdctlIgnore("network", "rm", podNetworkName)
}

var v1apispec *openapi3.T
var disco gostwire.DiscoveryResult

var _ = BeforeSuite(func() {
	if os.Getuid() != 0 {
		return
	}

	By("loading the v1 specification")
	var err error
	v1apispec, err = openapi3.NewLoader().LoadFromFile("../openapi-spec/ghostwire-v1.yaml")
	Expect(err).To(Succeed())
	Expect(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return v1apispec.Validate(ctx)
	}()).To(Succeed(), "Ghostwire API v1 OpenAPI specification is invalid")

	By("first making tabula rasa")
	tabulaRasa()

	By("setting up some fake pod containers")
	nerdctl.Nerdctl("network", "create", podNetworkName)
	nerdctl.Nerdctl(
		"run", "-d",
		"--name", bareName,
		"busybox", "/bin/sleep", "120s")
	nerdctl.Nerdctl(
		"run", "-d",
		"--name", podc1,
		"--network", podNetworkName,
		"--label", kuhbernetes.PodNamespaceLabel+"="+podNamespace,
		"--label", kuhbernetes.PodNameLabel+"="+podName,
		"--label", kuhbernetes.PodContainerNameLabel+"="+podc1,
		"busybox", "/bin/sleep", "120s")
	nerdctl.Nerdctl(
		"run", "-d",
		"--name", podc2,
		"--network", podNetworkName,
		"--label", kuhbernetes.PodNamespaceLabel+"="+podNamespace,
		"--label", kuhbernetes.PodNameLabel+"="+podName,
		"--label", kuhbernetes.PodContainerNameLabel+"="+podc2,
		"busybox", "/bin/sleep", "120s")

	By("discovering")
	ctx, cancel := context.WithCancel(context.Background())
	cizer := turtlefinder.New(func() context.Context { return ctx })
	defer cancel()
	defer cizer.Close()
	disco = gostwire.Discover(ctx, cizer, map[string]string{
		ieappicon.IEAppDiscoveryLabel: "",
	})
	Expect(disco.Lxkns.Containers).To(
		ContainElement(util.FromPod(podFQDN)),
	)
	for _, c := range disco.Lxkns.Containers {
		if c.Name == bareName {
			c.Labels[turtlefinder.TurtlefinderContainerPrefixLabelName] = barePrefix
			break
		}
	}

	DeferCleanup(func() {
		By("cleaning up afterwards")
		tabulaRasa()
	})
})
