package api_test

import (
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/leadership-election/internal/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leader", func() {
	It("returns a 200 OK when /v1/leader is hit", func() {
		server := httptest.NewServer(http.HandlerFunc(api.LeaderHandler))
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/v1/leader", nil)

		resp, err := http.DefaultClient.Do(req)
		Expect(err).To(Not(HaveOccurred()))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})
