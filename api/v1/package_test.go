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
	"github.com/thediveo/lxkns/decorator/kuhbernetes"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/decorator/ieappicon"
	"github.com/siemens/ghostwire/v2/test/nerdctl"
	"github.com/siemens/ghostwire/v2/turtlefinder"
	"github.com/siemens/ghostwire/v2/util"

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
			c.Labels[turtlefinder.GostwireContainerPrefixLabelName] = barePrefix
			break
		}
	}

	DeferCleanup(func() {
		By("cleaning up afterwards")
		tabulaRasa()
	})
})
