package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"io/ioutil"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
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
	jsonFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("Failed opening config file", err.Error())
	}
	var config signalformConfig
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing config file", err.Error())
	}
	return &config, nil
}

type signalformConfig struct {
	SfxToken string `json:"sfx_token"`
}
