//go:build e2e

package e2e_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
)

type pfTopology string

const (
	samePF         pfTopology = "same-pf"
	twoPFs         pfTopology = "two-pfs"
	withTLBDynamicLb           = "tlb-dynamic-lb"
)

type bondTestConfig struct {
	mode       string
	hashPolicy string
	pfTopology pfTopology
}

func createSriovNetworkNodePolicy(ctx context.Context, pf pfTopology) {
	By(fmt.Sprintf("creating SR-IOV network node policy for %s", pf))
}

func deleteSriovNetworkNodePolicy(ctx context.Context) {
	By("deleting SR-IOV network node policy")
}

func deployBondNAD(ctx context.Context, mode, hashPolicy string) {
	By(fmt.Sprintf("deploying Bond NAD mode=%s hashPolicy=%s", mode, hashPolicy))
}

func deploySUTPod(ctx context.Context, pf pfTopology) {
	By(fmt.Sprintf("deploying SUT pod with %s", pf))
}

func deployTargetPods(ctx context.Context, pf pfTopology) {
	By(fmt.Sprintf("deploying target pods A, B, C for %s", pf))
}

func verifyBondInterfaces(ctx context.Context) {
	By("verifying bond interfaces net1, net2, bond0 are up")
}

func generateTraffic(ctx context.Context) {
	By("generating ICMP, TCP, UDP, Multicast traffic to target pods")
}

func verifyLoadBalancing(ctx context.Context) {
	By("verifying traffic is captured on both net1 and net2 via tcpdump")
}

func simulateLinkFailure(ctx context.Context, iface string) {
	By(fmt.Sprintf("simulating link failure: ip link set %s down", iface))
}

func verifyFailover(ctx context.Context) {
	By("verifying traffic continues on surviving interface after link failure")
}

func restoreLink(ctx context.Context, iface string) {
	By(fmt.Sprintf("restoring link: ip link set %s up", iface))
}

func runBondTest(ctx context.Context, cfg bondTestConfig) {
	deployBondNAD(ctx, cfg.mode, cfg.hashPolicy)
	deploySUTPod(ctx, cfg.pfTopology)
	deployTargetPods(ctx, cfg.pfTopology)
	verifyBondInterfaces(ctx)
	generateTraffic(ctx)
	verifyLoadBalancing(ctx)
	simulateLinkFailure(ctx, "net1")
	verifyFailover(ctx)
	restoreLink(ctx, "net1")
	verifyLoadBalancing(ctx)
}

var _ = Describe("Bond CNI", func() {

	Describe("balance-xor", func() { // switch: static LAG

		Context("same PF", func() {
			BeforeEach(func(ctx context.Context) {
				createSriovNetworkNodePolicy(ctx, samePF)
				DeferCleanup(deleteSriovNetworkNodePolicy)
			})

			It("distributes traffic with layer2 hash", func(ctx context.Context) {
				runBondTest(ctx, bondTestConfig{"balance-xor", "layer2", samePF})
			})
		})

		Context("two PFs", func() {
			BeforeEach(func(ctx context.Context) {
				createSriovNetworkNodePolicy(ctx, twoPFs)
				DeferCleanup(deleteSriovNetworkNodePolicy)
			})

			DescribeTable("distributes traffic",
				func(ctx context.Context, hashPolicy string) {
					runBondTest(ctx, bondTestConfig{"balance-xor", hashPolicy, twoPFs})
				},
				Entry("with layer2 hash", "layer2"),
				Entry("with layer3+4 hash", "layer3+4"),
			)
		})
	})

	Describe("balance-tlb", func() { // switch: autonomous

		Context("same PF", func() {
			BeforeEach(func(ctx context.Context) {
				createSriovNetworkNodePolicy(ctx, samePF)
				DeferCleanup(deleteSriovNetworkNodePolicy)
			})

			It("distributes traffic with tlbDynamicLb", func(ctx context.Context) {
				runBondTest(ctx, bondTestConfig{"balance-tlb", withTLBDynamicLb, samePF})
			})
		})

		Context("two PFs", func() {
			BeforeEach(func(ctx context.Context) {
				createSriovNetworkNodePolicy(ctx, twoPFs)
				DeferCleanup(deleteSriovNetworkNodePolicy)
			})

			It("distributes traffic with tlbDynamicLb", Label("pending-dev-approval"), func(ctx context.Context) {
				runBondTest(ctx, bondTestConfig{"balance-tlb", withTLBDynamicLb, twoPFs})
			})
		})
	})

	Describe("802.3ad", func() { // switch: LACP

		Context("two PFs", func() {
			BeforeEach(func(ctx context.Context) {
				createSriovNetworkNodePolicy(ctx, twoPFs)
				DeferCleanup(deleteSriovNetworkNodePolicy)
			})

			DescribeTable("distributes traffic",
				func(ctx context.Context, hashPolicy string) {
					runBondTest(ctx, bondTestConfig{"802.3ad", hashPolicy, twoPFs})
				},
				Entry("with layer2 hash", "layer2"),
				Entry("with layer3+4 hash", "layer3+4"),
			)
		})
	})
})
