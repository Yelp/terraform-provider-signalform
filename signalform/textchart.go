package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
)

func textchartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Setting synced to 1 implies that the detector in SignalForm and SignalFx are identical",
			},
			"last_updated": &schema.Schema{
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Latest timestamp the resource was updated",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the chart",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the chart (Optional)",
			},
			"markdown": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Markdown text to display. More info at: \"https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet\"",
			},
		},

		Create: textchartCreate,
		Read:   textchartRead,
		Update: textchartUpdate,
		Delete: textchartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a text chart
*/
func getPayloadTextChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
	}

	viz := getTextChartOptions(d)
	if len(viz) > 0 {
		payload["options"] = viz
	}

	a, e := json.Marshal(payload)
	_ = ioutil.WriteFile("/tmp/fdc_chartCreate", a, 0644)

	return a, e
}

func getTextChartOptions(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "Text"
	if val, ok := d.GetOk("markdown"); ok {
		viz["markdown"] = val.(string)
	}

	return viz
}

func textchartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadTextChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.SfxToken, payload, d)
}

func textchartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.SfxToken, d)
}

func textchartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadTextChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.SfxToken, payload, d)
}

func textchartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())
	return resourceDelete(url, config.SfxToken, d)
}
