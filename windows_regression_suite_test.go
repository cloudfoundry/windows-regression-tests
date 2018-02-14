package windows_regression_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWindowsRegression(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WindowsRegression Suite")
}
