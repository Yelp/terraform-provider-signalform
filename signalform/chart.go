package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"strings"
)

func chartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"last_updated": &schema.Schema{
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"programText": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"chart_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"unit_prefix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"color_by": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"show_event_lines": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"stacked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_plot_type": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"minimum_resolution": &schema.Schema{
				Type: schema.TypeInt,
				// TODO: not sure about this
				Optional: true,
			},
			"max_delay": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"disable_sampling": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		Create: chartCreate,
		Read:   chartRead,
		Update: chartUpdate,
		Delete: chartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a chart
*/
func getPayloadChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("programText").(string),
	}

	if viz := getVisualizationOptionsChart(d); len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
}

func getVisualizationOptionsChart(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	if val, ok := d.GetOk("chart_type"); ok {
		viz["type"] = val.(string)
	}
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		viz["colorBy"] = val.(string)
	}
	if val, ok := d.GetOk("show_event_lines"); ok {
		viz["showEventLines"] = val.(bool)
	}
	if val, ok := d.GetOk("stacked"); ok {
		viz["stacked"] = val.(bool)
	}
	if val, ok := d.GetOk("default_plot_type"); ok {
		viz["defaultPlotType"] = val.(string)
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("minimum_resolution"); ok {
		programOptions["minimumResolution"] = val.(int)
	}
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int)
	}
	if val, ok := d.GetOk("disable_sampling"); ok {
		programOptions["disableSampling"] = val.(bool)
	}
	if len(programOptions) > 0 {
		viz["programOptions"] = programOptions
	}

	timeMap := make(map[string]interface{})
	if val, ok := d.GetOk("time_span_type"); ok {
		timeMap["type"] = val.(string)
	}
	if val, ok := d.GetOk("time_range"); ok {
		timeMap["range"] = val.(int)
	}
	if val, ok := d.GetOk("start_time"); ok {
		timeMap["start"] = val.(int)
	}
	if val, ok := d.GetOk("end_time"); ok {
		timeMap["end"] = val.(int)
	}
	if len(timeMap) > 0 {
		viz["time"] = timeMap
	}

	axisOptions := make(map[string]interface{})
	if val, ok := d.GetOk("min_value_axis"); ok {
		axisOptions["min"] = val.(int)
	}
	if val, ok := d.GetOk("max_value_axis"); ok {
		axisOptions["max"] = val.(int)
	}
	if val, ok := d.GetOk("label_axis"); ok {
		axisOptions["label"] = val.(string)
	}
	if val, ok := d.GetOk("line_high_watermark"); ok {
		axisOptions["highWatermark"] = val.(int)
	}
	if val, ok := d.GetOk("line_low_watermark"); ok {
		axisOptions["lowWatermark"] = val.(int)
	}
	if len(timeMap) > 0 {
		viz["axes"] = axisOptions
	}

	lineChartOptions := make(map[string]interface{})
	if val, ok := d.GetOk("metric_property"); ok {
		lineChartOptions["property"] = val.(string)
	}
	if val, ok := d.GetOk("display_metric_property"); ok {
		lineChartOptions["enabled"] = val.(bool)
	}
	if len(timeMap) > 0 {
		viz["legendOptions"] = lineChartOptions
	}

	areaChartOptions := make(map[string]interface{})
	if val, ok := d.GetOk("show_data_marker"); ok {
		areaChartOptions["showDataMarkers"] = val.(string)
	}
	if len(timeMap) > 0 {
		viz["areaChartOptions"] = lineChartOptions
	}
	return viz
}

/*
  Fetches payload specified in terraform configuration and creates chart
*/
func chartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := config.ProviderEndpoint
	payload, err := getPayloadChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	status_code, resp_body, err := sendRequest("POST", url, config.SfxToken, payload)
	_ = ioutil.WriteFile("/tmp/fdc_chartCreate", resp_body, 0644)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for chart %s during creation: %s", d.Get("name"), err.Error())
		}
		d.SetId(fmt.Sprintf("%s", mapped_resp["id"].(string)))
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
	} else {
		return fmt.Errorf("For chart %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Send a GET to get the current state of the chart.  It just checks if the lastUpdated timestamp is
  later than the timestamp saved in the resource.  If so, the chart has been modified in some way
  in the UI, and should be recreated.  This is signaled by setting synced to 0, meaning if synced is set to
  1 in the tf configuration, it will update the chart to achieve the desired state.
*/
func chartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())

	status_code, resp_body, err := sendRequest("GET", url, config.SfxToken, nil)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for chart %s during read: %s", d.Get("name"), err.Error())
		}
		// This implies the chart was modified in the Signalfx UI and therefore it is not synced with Signalform
		last_updated := mapped_resp["lastUpdated"].(float64)
		if last_updated > (d.Get("last_updated").(float64) + OFFSET) {
			d.Set("synced", 0)
			d.Set("last_updated", last_updated)
		}
	} else {
		if strings.Contains(string(resp_body), "Chart not found") {
			// This implies chart was deleted in the Signalfx UI and therefore we need to recreate it
			d.SetId("")
		} else {
			return fmt.Errorf("For Chart %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
		}
	}
	return nil
}

func chartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())

	status_code, resp_body, err := sendRequest("PUT", url, config.SfxToken, payload)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for chart %s during creation: %s", d.Get("name"), err.Error())
		}
		// If the chart was updated successfully with Signalform configs, it is now synced with Signalfx
		d.Set("synced", 1)
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
	} else {
		return fmt.Errorf("For Chart %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Deletes a chart.  If the chart does not exist, it will receive a 404, and carry on as usual.
*/
func chartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())
	status_code, resp_body, err := sendRequest("DELETE", url, config.SfxToken, nil)
	if err != nil {
		return fmt.Errorf("Failed deleting chart %s: %s", d.Get("name"), err.Error())
	}
	if status_code < 400 || status_code == 404 {
		d.SetId("")
	} else {
		return fmt.Errorf("For Chart %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}
