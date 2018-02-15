package windows_regression_tests

import (
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

var _ = Describe("Forkbomb", func() {
	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("this test may not pass on windows2016")
		}

		memLimit := config.GetNumWindowsCells() * 2 * 4
		if memLimit < 10 {
			memLimit = 10
		}
		setTotalMemoryLimit(fmt.Sprintf("%dG", memLimit))
	})

	AfterEach(func() {
		setTotalMemoryLimit("10G")
	})

	It("cannot forkbomb the environment", func() {
		pushNoraWithOptions(appName, config.GetNumWindowsCells()*2, "2G")

		Eventually(allAppInstancesRunning(appName, config.GetNumWindowsCells()*2, CF_PUSH_TIMEOUT), CF_PUSH_TIMEOUT).Should(Succeed())

		computerNames := reportedComputerNames(config.GetNumWindowsCells())
		Expect(len(computerNames)).To(Equal(config.GetNumWindowsCells()))

		helpers.CurlApp(config, appName, "/run", "-f", "-X", "POST", "-d", "bin/breakoutbomb.exe")

		time.Sleep(3 * time.Second)

		newComputerNames := reportedComputerNames(config.GetNumWindowsCells())
		Expect(newComputerNames).To(Equal(computerNames))
	})
})

func setTotalMemoryLimit(memoryLimit string) {
	type quotaDefinitionUrl struct {
		Resources []struct {
			Entity struct {
				QuotaDefinitionUrl string `json:"quota_definition_url"`
			} `json:"entity"`
		} `json:"resources"`
	}

	orgEndpoint := fmt.Sprintf("/v2/organizations?q=name%%3A%s", environment.GetOrganizationName())
	var org quotaDefinitionUrl
	workflowhelpers.ApiRequest("GET", orgEndpoint, &org, DEFAULT_TIMEOUT)
	Expect(org.Resources).ToNot(BeEmpty())

	type quotaDefinition struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	var quota quotaDefinition
	workflowhelpers.ApiRequest("GET", org.Resources[0].Entity.QuotaDefinitionUrl, &quota, DEFAULT_TIMEOUT)
	Expect(quota.Entity.Name).ToNot(BeEmpty())

	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("update-quota", quota.Entity.Name, "-m", memoryLimit).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

func allAppInstancesRunning(appName string, instances int, timeout time.Duration) func() error {
	return func() error {
		type StatsResponse map[string]struct {
			State string `json:"state"`
		}

		buf, err := runCfWithOutput("app", appName, "--guid")
		if err != nil {
			return err
		}
		appGuid := strings.Replace(string(buf.Contents()), "\n", "", -1)

		endpoint := fmt.Sprintf("/v2/apps/%s/stats", appGuid)

		var response StatsResponse
		workflowhelpers.ApiRequest("GET", endpoint, &response, timeout)

		err = nil
		for k, v := range response {
			if v.State != "RUNNING" {
				err = errors.New(fmt.Sprintf("App %s instance %s is not running: State = %s", appName, k, v.State))
			}
		}
		return err
	}
}

func reportedComputerNames(instances int) map[string]bool {
	timer := time.NewTimer(time.Second * 120)
	defer timer.Stop()
	run := true
	go func() {
		<-timer.C
		run = false
	}()

	seenComputerNames := map[string]bool{}
	for len(seenComputerNames) != instances && run == true {
		seenComputerNames[helpers.CurlApp(config, appName, "/ENV/CF_INSTANCE_IP")] = true
	}

	return seenComputerNames
}
