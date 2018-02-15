package windows_regression_tests

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

const (
	DEFAULT_TIMEOUT      = 45 * time.Second
	DEFAULT_LONG_TIMEOUT = 4 * DEFAULT_TIMEOUT
	CF_PUSH_TIMEOUT      = 3 * time.Minute
)

var (
	appName     string
	config      *watsConfig
	environment *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestCFWindows(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	var err error
	config, err = loadWatsConfig()
	if err != nil {
		t.Fatalf("could not load WATS config: %s", err.Error())
	}

	environment = workflowhelpers.NewTestSuiteSetup(config)

	BeforeSuite(func() {
		environment.Setup()

		if config.GetIsolationSegmentName() != "" {
			workflowhelpers.AsUser(environment.AdminUserContext(), environment.ShortTimeout(), func() {
				isoSegGuid := createOrGetIsolationSegment(config.GetIsolationSegmentName())
				attachIsolationSegmentToOrg(environment, isoSegGuid)
				attachIsolationSegmentToSpace(environment, isoSegGuid)
			})
		}
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	BeforeEach(func() {
		Eventually(cf.Cf("apps").Out).Should(gbytes.Say("No apps found"))
		appName = generator.PrefixedRandomName(config.GetNamePrefix(), "APP")
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(gexec.Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(gexec.Exit(0))
	})

	componentName := "CF Windows"

	rs := []Reporter{}

	if config.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(config, componentName)
		rs = append(rs, helpers.NewJUnitReporter(config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

func createOrGetIsolationSegment(isolationSegmentName string) string {
	// This could go in cf-test-helpers
	guid := getIsolationSegmentGuid(isolationSegmentName)
	if guid == "" {
		Eventually(cf.Cf("curl", "/v3/isolation_segments/", "-X", "POST", "-d", fmt.Sprintf(`{"name":"%s"}`, isolationSegmentName))).Should(gexec.Exit(0))
		guid = getIsolationSegmentGuid(isolationSegmentName)
	}
	return guid
}

func getIsolationSegmentGuid(isolationSegmentName string) string {
	return getV3ResourceGuid(fmt.Sprintf("/v3/isolation_segments?names=%s", isolationSegmentName))
}

func attachIsolationSegmentToOrg(environment *workflowhelpers.ReproducibleTestSuiteSetup, isoSegGuid string) {
	orgGuid := getOrganizationGuid(environment.GetOrganizationName())
	response := cf.Cf(
		"curl",
		fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isoSegGuid),
		"-X",
		"POST",
		"-d",
		fmt.Sprintf(`{"data":[{"guid": "%s"}]}`, orgGuid),
	)
	Expect(response.Wait()).To(gexec.Exit(0))
}

func attachIsolationSegmentToSpace(environment *workflowhelpers.ReproducibleTestSuiteSetup, isoSegGuid string) {
	spaceGuid := getSpaceGuidForOrg(getOrganizationGuid(environment.GetOrganizationName()))
	response := cf.Cf(
		"curl",
		fmt.Sprintf("/v2/spaces/%s", spaceGuid),
		"-X",
		"PUT",
		"-d",
		fmt.Sprintf(`{"isolation_segment_guid":"%s"}`, isoSegGuid),
	)
	Expect(response.Wait()).To(gexec.Exit(0))
}

func getOrganizationGuid(organizationName string) string {
	return getV2ResourceGuid(fmt.Sprintf("/v2/organizations?q=name:%s", organizationName))
}

func getSpaceGuidForOrg(orgGuid string) string {
	return getV2ResourceGuid(fmt.Sprintf("/v2/organizations/%s/spaces", orgGuid))
}

func getV2ResourceGuid(endpoint string) string {
	response := cf.Cf("curl", endpoint, "-X", "GET")
	Expect(response.Wait()).To(gexec.Exit(0))
	var r struct {
		Resources []struct {
			Metadata struct {
				Guid string
			}
		}
	}
	Expect(json.Unmarshal(response.Out.Contents(), &r)).To(Succeed())
	Expect(len(r.Resources)).To(Equal(1))
	return r.Resources[0].Metadata.Guid
}

func getV3ResourceGuid(endpoint string) string {
	response := cf.Cf("curl", endpoint, "-X", "GET")
	Expect(response.Wait()).To(gexec.Exit(0))
	var r struct {
		Resources []struct {
			Guid string
		}
	}
	Expect(json.Unmarshal(response.Out.Contents(), &r)).To(Succeed())
	if len(r.Resources) == 0 {
		return ""
	}
	return r.Resources[0].Guid
}

func runCfWithOutput(values ...string) (*gbytes.Buffer, error) {
	session := cf.Cf(values...)
	session.Wait(CF_PUSH_TIMEOUT)
	if session.ExitCode() == 0 {
		return session.Out, nil
	}

	return session.Out, fmt.Errorf("non zero exit code %d", session.ExitCode())
}

func pushNora(appName string) {
	By("pushing it")
	ExpectWithOffset(1, pushApp(appName, "./assets/nora/NoraPublished", 1, "256M", "hwc_buildpack").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

	By("verifying it's up")
	EventuallyWithOffset(1, helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
}

func pushNoraWithOptions(appName string, instances int, memory string) {
	By("pushing it")
	ExpectWithOffset(1, pushApp(appName, "./assets/nora/NoraPublished", instances, memory, "hwc_buildpack").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

	By("verifying it's up")
	EventuallyWithOffset(1, helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
}

func pushApp(appName, path string, instances int, memory, buildpack string, args ...string) *gexec.Session {
	cfArgs := []string{
		"push", appName,
		"-p", path,
		"-i", strconv.Itoa(instances),
		"-m", memory,
		"-b", buildpack,
		"-s", config.GetStack(),
	}
	cfArgs = append(cfArgs, args...)
	return cf.Cf(cfArgs...)
}
