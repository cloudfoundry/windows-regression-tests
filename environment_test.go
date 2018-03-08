package windows_regression_tests

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application environment", func() {
	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("n/a on windows2016")
		}
	})

	It("should not have too many environment variable exposed", func() {
		pushNora(appName)

		excludedList := []string{
			"COMPUTERNAME",
			"ALLUSERSPROFILE",
			"FP_NO_HOST_CHECK",
			"GOPATH",
			"NUMBER_OF_PROCESSORS",
			"OS",
			"PROCESSOR_ARCHITECTURE",
			"PROCESSOR_IDENTIFIER",
			"PROCESSOR_LEVEL",
			"PROCESSOR_REVISION",
			"PSModulePath",
			"PUBLIC",
			"SystemDrive",
			"USERDOMAIN",
			"VS110COMNTOOLS",
			"VS120COMNTOOLS",
			"WIX",
		}
		response := helpers.CurlApp(config, appName, "/env")
		var env map[string]string
		json.Unmarshal([]byte(response), &env)
		for _, excludedKey := range excludedList {
			Expect(env).NotTo(HaveKey(excludedKey))
		}
	})
})
