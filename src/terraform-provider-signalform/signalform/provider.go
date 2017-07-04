package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"io/ioutil"
	"os"
	"os/user"
)

var SystemConfigPath = "/etc/signalfx.conf"
var HomeConfigSuffix = "/.signalfx.conf"
var HomeConfigPath = ""

type signalformConfig struct {
	AuthToken string `json:"auth_token"`
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SFX_AUTH_TOKEN", ""),
				Description: "SignalFx auth token",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"signalform_detector":           detectorResource(),
			"signalform_time_chart":         timeChartResource(),
			"signalform_heatmap_chart":      heatmapChartResource(),
			"signalform_single_value_chart": singleValueChartResource(),
			"signalform_list_chart":         listChartResource(),
			"signalform_text_chart":         textChartResource(),
			"signalform_dashboard":          dashboardResource(),
			"signalform_dashboard_group":    dashboardGroupResource(),
		},
		ConfigureFunc: signalformConfigure,
	}
}

func signalformConfigure(data *schema.ResourceData) (interface{}, error) {
	config := signalformConfig{}

	// /etc/signalfx.conf has lowest priority
	if _, err := os.Stat(SystemConfigPath); err == nil {
		err = readConfigFile(SystemConfigPath, &config)
		if err != nil {
			return nil, err
		}
	}

	// $HOME/.signalfx.conf second
	// this additional variable is used for mocking purposes in tests
	if HomeConfigPath == "" {
		usr, err := user.Current()
		HomeConfigPath = usr.HomeDir + HomeConfigSuffix
		if err != nil {
			return nil, fmt.Errorf("Failed to get user environment %s", err.Error())
		}
	}
	if _, err := os.Stat(HomeConfigPath); err == nil {
		err = readConfigFile(HomeConfigPath, &config)
		if err != nil {
			return nil, err
		}
	}

	// provider first
	if token, ok := data.GetOk("auth_token"); ok {
		config.AuthToken = token.(string)
	}

	if config.AuthToken == "" {
		return &config, fmt.Errorf("auth_token: required field is not set")
	}

	return &config, nil

}

func readConfigFile(configPath string, config *signalformConfig) error {
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Failed to open config file. %s", err.Error())
	}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return fmt.Errorf("Failed to parse config file. %s", err.Error())
	}
	return nil
}
