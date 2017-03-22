package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"math"
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the chart",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the chart (Optional)",
			},
			"program_text": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Signalflow program text for the chart. More info at \"https://developers.signalfx.com/docs/signalflow-overview\"",
			},
			"unit_prefix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "(Metric by default) Must be \"Metric\" or \"Binary\"",
			},
			"color_by": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "(Dimension by default) Must be \"Dimension\" or \"Metric\"",
			},
			"minimum_resolution": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum resolution to use for computing the underlying program",
			},
			"max_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How long to wait for late datapoints",
			},
			"disable_sampling": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) If false, samples a subset of the output MTS, which improves UI performance",
			},
			"time_span_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Type of time interval of the chart. It must be \"absolute\" or \"relative\"",
				ValidateFunc: validateTimeSpanType,
			},
			"time_range": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "(time_span_type \"relative\" only) Absolute minutes offset from now to visualize",
				ConflictsWith: []string{"start_time", "end_time"},
			},
			"start_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "(type \"absolute\" only) Seconds since epoch to start the visualization",
				ConflictsWith: []string{"time_range"},
			},
			"end_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "(type \"absolute\" only) Seconds since epoch to end the visualization",
				ConflictsWith: []string{"time_range"},
			},
			"axis_left": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_value": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     math.MinInt32,
							Description: "The minimum value for the left axis",
						},
						"max_value": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     math.MaxInt32,
							Description: "The maximum value for the left axis",
						},
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Label of the left axis",
						},
						"high_watermark": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     math.MaxInt32,
							Description: "A line to draw as a high watermark",
						},
						"low_watermark": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     math.MinInt32,
							Description: "A line to draw as a low watermark",
						},
					},
				},
			},
			"properties_to_hide": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of properties that shouldn't be displayed in the chart legend (i.e. dimension names)",
			},
			"show_event_lines": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether vertical highlight lines should be drawn in the visualizations at times when events occurred",
			},
			"line_show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Show markers (circles) for each datapoint used to draw line charts",
			},
			"area_show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Show markers (circles) for each datapoint used to draw area charts",
			},
			"stacked": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether area and bar charts in the visualization should be stacked",
			},
			"plot_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "(LineChart by default) The default plot display style for the visualization. Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"",
				ValidateFunc: validatePlotTypeTimeChart,
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

	viz := getVisualizationOptionsChart(d)
	//	if viz2 := getLineChartOptions(d); len(viz2) > 0 {
	//		viz["fields"] = viz2
	//	}
	if axesOptions := getAxesOptions(d); len(axesOptions) > 0 {
		viz["axes"] = axesOptions
	}
	if legendOptions := getLegendOptions(d); len(legendOptions) > 0 {
		viz["legendOptions"] = legendOptions
	}
	if len(viz) > 0 {
		payload["options"] = viz
	}

	a, e := json.Marshal(payload)
	return a, e
}

func getLegendOptions(d *schema.ResourceData) map[string]interface{} {
	if properties, ok := d.GetOk("properties_to_hide"); ok {
		properties := properties.([]interface{})
		legendOptions := make(map[string]interface{})
		properties_opts := make([]map[string]interface{}, len(properties))
		for i, property := range properties {
			property := property.(string)
			item := make(map[string]interface{})
			item["property"] = property
			item["enabled"] = false
			properties_opts[i] = item
		}
		if len(properties_opts) > 0 {
			legendOptions["fields"] = properties_opts
			return legendOptions
		}
	}
	return nil
}

func getAxesOptions(d *schema.ResourceData) []map[string]interface{} {
	if tf_axis_opts, ok := d.GetOk("axis_left"); ok {
		tf_left_axis_opts := tf_axis_opts.(*schema.Set).List()
		axes_list_opts := make([]map[string]interface{}, len(tf_left_axis_opts))
		for i, tf_opt := range tf_left_axis_opts {
			tf_opt := tf_opt.(map[string]interface{})
			item := make(map[string]interface{})

			if val, ok := tf_opt["min_value"]; ok {
				if val.(int) == math.MinInt32 {
					item["min"] = nil
				} else {
					item["min"] = val.(int)
				}
			}
			if val, ok := tf_opt["max_value"]; ok {
				if val.(int) == math.MaxInt32 {
					item["max"] = nil
				} else {
					item["max"] = val.(int)
				}
			}
			if val, ok := tf_opt["label"]; ok {
				item["label"] = val.(string)
			}
			if val, ok := tf_opt["high_watermark"]; ok {
				if val.(int) == math.MaxInt32 {
					item["highWatermark"] = nil
				} else {
					item["highWatermark"] = val.(int)
				}
			}
			if val, ok := tf_opt["low_watermark"]; ok {
				if val.(int) == math.MinInt32 {
					item["lowWatermark"] = nil
				} else {
					item["lowWatermark"] = val.(int)
				}
			}
			axes_list_opts[i] = item
		}
		return axes_list_opts
	}
	return nil
}

func getVisualizationOptionsChart(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "TimeSeriesChart"
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
		timeMap["range"] = val.(int) * 60 * 1000
	}
	if val, ok := d.GetOk("start_time"); ok {
		timeMap["start"] = val.(int) * 1000
	}
	if val, ok := d.GetOk("end_time"); ok {
		timeMap["end"] = val.(int) * 1000
	}
	if len(timeMap) > 0 {
		viz["time"] = timeMap
	}

	areaChartOptions := make(map[string]interface{})
	if val, ok := d.GetOk("area_show_data_markers"); ok {
		areaChartOptions["showDataMarkers"] = val.(bool)
	}
	if len(areaChartOptions) > 0 {
		viz["areaChartOptions"] = areaChartOptions
	}

	lineChartOptions := make(map[string]interface{})
	if val, ok := d.GetOk("line_show_data_markers"); ok {
		lineChartOptions["showDataMarkers"] = val.(bool)
	}
	if len(lineChartOptions) > 0 {
		viz["lineChartOptions"] = lineChartOptions
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
