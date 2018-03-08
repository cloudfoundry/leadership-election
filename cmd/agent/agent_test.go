package agent_test

import (
	"code.cloudfoundry.org/leadership-election/cmd/agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent", func() {
	It("identifies itself as the leader", func() {
		a := agent.NewAgent()
		Expect(a.IsLeader).To(BeTrue())
	})

})
