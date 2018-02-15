package windows_regression_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	CredhubAssistedMode    = "assisted"
	CredhubNonAssistedMode = "non-assisted"
)

type wartsConfig struct {
	ApiEndpoint          string `json:"api"`
	AdminUser            string `json:"admin_user"`
	AdminPassword        string `json:"admin_password"`
	AppsDomain           string `json:"apps_domain"`
	SkipSSLValidation    bool   `json:"skip_ssl_validation"`
	NumWindowsCells      int    `json:"num_windows_cells"`
	ArtifactsDirectory   string `json:"artifacts_directory"`
	UseHttp              bool   `json:"use_http"`
	IsolationSegmentName string `json:"isolation_segment_name"`
	Stack                string `json:"stack"`
}

func loadWartsConfig() (*wartsConfig, error) {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		return &wartsConfig{}, errors.New("Must set CONFIG to point to an integration config JSON file")
	}

	return loadWartsConfigFromPath(configPath)
}

func loadWartsConfigFromPath(configPath string) (*wartsConfig, error) {
	configContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := wartsConfig{
		ArtifactsDirectory: filepath.Join("..", "results"),
		UseHttp:            true,
	}
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	switch config.GetStack() {
	case "windows2012R2", "windows2016":
	default:
		return nil, errors.New("Invalid stack: " + config.GetStack())
	}

	if config.NumWindowsCells == 0 {
		return nil, fmt.Errorf("Please provide 'num_windows_cells' as a property in the integration config JSON (The number of windows cells in tested deployment)")
	}

	return &config, nil
}

func (w *wartsConfig) GetApiEndpoint() string {
	return w.ApiEndpoint
}

func (w *wartsConfig) GetConfigurableTestPassword() string {
	return ""
}

func (w *wartsConfig) GetPersistentAppOrg() string {
	return ""
}

func (w *wartsConfig) GetPersistentAppQuotaName() string {
	return ""
}

func (w *wartsConfig) GetPersistentAppSpace() string {
	return ""
}

func (w *wartsConfig) GetScaledTimeout(timeout time.Duration) time.Duration {
	return timeout
}

func (w *wartsConfig) GetAdminPassword() string {
	return w.AdminPassword
}

func (w *wartsConfig) GetExistingUser() string {
	return ""
}

func (w *wartsConfig) GetExistingUserPassword() string {
	return ""
}

func (w *wartsConfig) GetShouldKeepUser() bool {
	return false
}

func (w *wartsConfig) GetUseExistingUser() bool {
	return false
}

func (w *wartsConfig) GetAdminUser() string {
	return w.AdminUser
}

func (w *wartsConfig) GetSkipSSLValidation() bool {
	return w.SkipSSLValidation
}

func (w *wartsConfig) GetNamePrefix() string {
	return "WARTS"
}

func (w *wartsConfig) GetAppsDomain() string {
	return w.AppsDomain
}

func (w *wartsConfig) GetNumWindowsCells() int {
	return w.NumWindowsCells
}

func (w *wartsConfig) GetArtifactsDirectory() string {
	return w.ArtifactsDirectory
}

func (w *wartsConfig) Protocol() string {
	if w.UseHttp {
		return "http://"
	} else {
		return "https://"
	}
}

func (w *wartsConfig) GetIsolationSegmentName() string {
	return w.IsolationSegmentName
}

func (w *wartsConfig) GetStack() string {
	return w.Stack
}

func (w *wartsConfig) GetUseExistingOrganization() bool {
	return false
}

func (w *wartsConfig) GetExistingOrganization() string {
	return ""
}

func (w *wartsConfig) GetUseExistingSpace() bool {
	return false
}

func (w *wartsConfig) GetExistingSpace() string {
	return ""
}
