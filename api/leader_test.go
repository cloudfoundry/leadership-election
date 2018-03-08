package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/leadership-election/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leader", func() {
	It("returns a 200 OK when /leader is hit", func() {
		server := httptest.NewServer(http.HandlerFunc(api.LeaderHandler))
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/leader", nil)

		resp, err := http.DefaultClient.Do(req)
		Expect(err).To(Not(HaveOccurred()))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).To(Not(HaveOccurred()))
		Expect(string(body)).To(Equal("true\n"))
	})
})
