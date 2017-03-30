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
			"detector":                    detectorResource(),
			"signalform_timechart":        timechartResource(),
			"signalform_heatmapchart":     heatmapchartResource(),
			"signalform_singlevaluechart": singlevaluechartResource(),
			"signalform_listchart":        listchartResource(),
			"signalform_textchart":        textchartResource(),
			"signalform_dashboard":        dashboardResource(),
			"signalform_dashboardgroup":   dashboardgroupResource(),
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
