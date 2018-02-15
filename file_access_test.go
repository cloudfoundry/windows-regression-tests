package windows_regression_tests

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File ACLs", func() {
	var inaccessibleFiles = []string{
		"C:\\bosh",
		"C:\\containerizer",
		"C:\\var",
		"C:\\Windows\\Panther\\Unattend",
	}

	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("n/a on windows2016")
		}
	})

	It("A Container user should not have permission to view sensitive files", func() {
		pushNora(appName)

		for _, path := range inaccessibleFiles {
			response, err := getFilePermission(path)
			Expect(err).To(Succeed(), path)

			response, err = strconv.Unquote(response)
			Expect(err).To(Succeed())
			Expect(response).To(Or(Equal("ACCESS_DENIED"), Equal("NOT_EXIST")), path)
		}
	})
})

func getFilePermission(path string) (string, error) {
	uri := helpers.AppUri(appName, "/inaccessible_file", config)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Post(uri, "text/plain", strings.NewReader(path))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return string(body), err
}
