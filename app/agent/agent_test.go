package agent_test

import (
	"fmt"
	"log"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/leadership-election/app/agent"
)

var run int = 10000

var _ = Describe("Agent", func() {
	var (
		a map[string]*agent.Agent
	)

	BeforeEach(func() {
		a = make(map[string]*agent.Agent)

		for i := 0; i < 3; i++ {
			g := agent.New(
				i,
				[]string{
					// Intra network addresses
					fmt.Sprintf("127.0.0.1:%d", run+3),
					fmt.Sprintf("127.0.0.1:%d", run+4),
					fmt.Sprintf("127.0.0.1:%d", run+5),

					// Fake and not serviced
					fmt.Sprintf("127.0.0.1:%d", run+6),
					fmt.Sprintf("127.0.0.1:%d", run+7),
					fmt.Sprintf("127.0.0.1:%d", run+8),
					fmt.Sprintf("127.0.0.1:%d", run+9),
					fmt.Sprintf("127.0.0.1:%d", run+10),
					fmt.Sprintf("127.0.0.1:%d", run+11),
					fmt.Sprintf("127.0.0.1:%d", run+12),
				},

				// External address
				agent.WithPort(run+i),
				agent.WithLogger(log.New(GinkgoWriter, fmt.Sprintf("[AGENT %d]", i), log.LstdFlags)),
			)
			g.Start()
			u := fmt.Sprintf("http://%s/v1/leader", g.Addr())

			a[u] = g
		}
	})

	AfterEach(func() {
		run += 2 * len(a)

		// We set up fake and not serviced addresses
		run += 7
	})

	It("returns a 200 if it is the leader", func() {
		var c []chan int

		for addr := range a {
			cc := make(chan int)
			c = append(c, cc)
			go func(addr string, cc chan int) {
				defer GinkgoRecover()
				for {
					resp, err := http.Get(addr)
					Expect(err).ToNot(HaveOccurred())

					if resp.StatusCode == http.StatusOK {
						cc <- 1
					}

					if resp.StatusCode == http.StatusLocked {
						cc <- 0
					}

					Expect(resp.StatusCode).To(Or(Equal(http.StatusOK), Equal(http.StatusLocked)))
				}
			}(addr, cc)
		}

		Eventually(func() int {
			var l0, l1, l2 int

			Eventually(c[0]).Should(Receive(&l0))
			Eventually(c[1]).Should(Receive(&l1))
			Eventually(c[2]).Should(Receive(&l2))

			return l0 + l1 + l2
		}, 10).Should(Equal(1))

		Consistently(func() int {
			var l0, l1, l2 int

			Eventually(c[0]).Should(Receive(&l0))
			Eventually(c[1]).Should(Receive(&l1))
			Eventually(c[2]).Should(Receive(&l2))

			return l0 + l1 + l2
		}, 3).Should(Equal(1))
	})
})
