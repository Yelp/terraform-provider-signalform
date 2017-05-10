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

type signalformConfig struct {
	SfxToken string `json:"auth_token"`
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SFX_AUTH_TOKEN", nil),
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
	// environment first
	if token, ok := data.GetOk("auth_token"); ok {
		return &signalformConfig{SfxToken: token.(string)}, nil
	}

	// $HOME/.signalfx.conf second
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("Failed to get user environment", err.Error())
	}
	configPath := usr.HomeDir + "/.signalfx.conf"
	if _, err := os.Stat(configPath); err == nil {
		return readConfigFile(configPath)
	}

	// /etc/signalfx.conf last
	return readConfigFile("/etc/signalfx.conf")
}

func readConfigFile(configPath string) (interface{}, error) {
	var config signalformConfig
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed opening config file. ", err.Error())
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing config file. ", err.Error())
	}
	return &config, nil
}
