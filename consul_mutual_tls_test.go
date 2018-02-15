package windows_regression_tests

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consul Mutual TLS", func() {
	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("n/a on windows2016")
		}
	})

	It("access to consul should be blocked", func() {
		pushNora(appName)
		response := helpers.CurlApp(config, appName, "/curl/127.0.0.1/8500")
		Expect(response).To(ContainSubstring("The server committed a protocol violation. Section=ResponseStatusLine"))
	})
})
