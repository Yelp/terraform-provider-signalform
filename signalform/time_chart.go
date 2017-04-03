package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"math"
)

func timeChartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Default:     true,
				Description: "Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing.",
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
				Description: "Description of the chart",
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
				Description: "The minimum resolution (in seconds) to use for computing the underlying program",
			},
			"max_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How long (in seconds) to wait for late datapoints",
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
			// TODO: Do the same for the axis_right as soon as signalfx relase the ability to
			// choose visualitation options at a metric level (since you can enable the right
			// axis at a metric level)
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
			"legend_fields_to_hide": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of properties that shouldn't be displayed in the chart legend (i.e. dimension names)",
			},
			"show_event_lines": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether vertical highlight lines should be drawn in the visualizations at times when events occurred",
			},
			"show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Show markers (circles) for each datapoint used to draw line or area charts",
			},
			"stacked": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Whether area and bar charts in the visualization should be stacked",
			},
			"plot_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "(LineChart by default) The default plot display style for the visualization. Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"",
				ValidateFunc: validatePlotTypeTimeChart,
			},
		},

		Create: timechartCreate,
		Read:   timechartRead,
		Update: timechartUpdate,
		Delete: timechartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a time chart
*/
func getPayloadTimeChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("program_text").(string),
	}

	viz := getTimeChartOptions(d)
	if axesOptions := getAxesOptions(d); len(axesOptions) > 0 {
		viz["axes"] = axesOptions
	}
	if legendOptions := getLegendOptions(d); len(legendOptions) > 0 {
		viz["legendOptions"] = legendOptions
	}
	if len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
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

func getTimeChartOptions(d *schema.ResourceData) map[string]interface{} {
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
	viz["stacked"] = d.Get("stacked").(bool)
	if val, ok := d.GetOk("plot_type"); ok {
		viz["defaultPlotType"] = val.(string)
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("minimum_resolution"); ok {
		programOptions["minimumResolution"] = val.(int) * 1000
	}
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int) * 1000
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

	dataMarkersOption := make(map[string]interface{})
	dataMarkersOption["showDataMarkers"] = d.Get("show_data_markers").(bool)
	if chartType, ok := d.GetOk("plot_type"); ok {
		chartType := chartType.(string)
		if chartType == "AreaChart" {
			viz["areaChartOptions"] = dataMarkersOption
		} else if chartType == "LineChart" {
			viz["lineChartOptions"] = dataMarkersOption
		}
	} else {
		viz["lineChartOptions"] = dataMarkersOption
	}

	return viz
}

func timechartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadTimeChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.SfxToken, payload, d)
}

func timechartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.SfxToken, d)
}

func timechartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadTimeChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.SfxToken, payload, d)
}

func timechartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())
	return resourceDelete(url, config.SfxToken, d)
}

/*
  Validates the plot_type field against a list of allowed words.
*/
func validatePlotTypeTimeChart(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "LineChart" && value != "AreaChart" && value != "ColumnChart" && value != "Histogram" {
		errors = append(errors, fmt.Errorf("%s not allowed; Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"", value))
	}
	return
}
