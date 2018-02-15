package windows_regression_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
)

var _ = Describe("ICMP Security Groups", func() {
	var icmpGroupName string

	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("n/a on windows2016")
		}

		rule := destination{IP: "0.0.0.0/0", Protocol: "icmp", Code: -1, Type: -1}
		icmpGroupName = createSecurityGroup(rule)
		bindSecurityGroup(icmpGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)
	})

	AfterEach(func() {
		unbindSecurityGroup(icmpGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)
		deleteSecurityGroup(icmpGroupName)
	})

	It("ignores the rule and can push an app", func() {
		pushNora(appName)
	})
})

type destination struct {
	IP       string `json:"destination"`
	Port     string `json:"ports,omitempty"`
	Protocol string `json:"protocol"`
	Code     int    `json:"code,omitempty"`
	Type     int    `json:"type,omitempty"`
}

func createSecurityGroup(allowedDestinations ...destination) string {
	file, _ := ioutil.TempFile(os.TempDir(), "WATS-sg-rules")
	defer os.Remove(file.Name())
	Expect(json.NewEncoder(file).Encode(allowedDestinations)).To(Succeed())

	rulesPath := file.Name()
	securityGroupName := fmt.Sprintf("WATS-SG-%s", generator.PrefixedRandomName(config.GetNamePrefix(), "SECURITY-GROUP"))

	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("create-security-group", securityGroupName, rulesPath).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})

	return securityGroupName
}

func bindSecurityGroup(securityGroupName, orgName, spaceName string) {
	By("Applying security group")
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("bind-security-group", securityGroupName, orgName, spaceName).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

func unbindSecurityGroup(securityGroupName, orgName, spaceName string) {
	By("Unapplying security group")
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("unbind-security-group", securityGroupName, orgName, spaceName).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

func deleteSecurityGroup(securityGroupName string) {
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("delete-security-group", securityGroupName, "-f").Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}
