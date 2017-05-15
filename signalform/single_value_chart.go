package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func singleValueChartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
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
				Description: "(Metric by default) Must be \"Metric\", \"Dimension\", or \"Scale\". \"Scale\" maps to Color by Value in the UI",
			},
			"max_delay": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints",
				ValidateFunc: validateMaxDelayValue,
			},
			"refresh_interval": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How often (in seconds) to refresh the values of the list",
			},
			"max_precision": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum precision to for values displayed in the list",
			},
			"is_timestamp_hidden": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Whether to hide the timestamp in the chart",
			},
			"show_spark_line": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether to show a trend line below the current value",
				Default:     false,
			},
			"color_scale": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Values for each color in the range. Example: { thresholds : [80, 60, 40, 20, 0], inverted : true }",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"thresholds": &schema.Schema{
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "The thresholds to set for the color range being used. Values must be in descending order",
						},
						"inverted": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "(false by default) If false or omitted, values are red if they are above the highest specified value. If true, values are red if they are below the lowest specified value",
						},
					},
				},
			},
			"viz_options": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Plot-level customization options, associated with a publish statement",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The label used in the publish statement that displays the plot (metric time series data) you want to customize",
						},
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Color to use",
							ValidateFunc: validatePerSignalColor,
						},
					},
				},
			},
		},

		Create: singlevaluechartCreate,
		Read:   singlevaluechartRead,
		Update: singlevaluechartUpdate,
		Delete: singlevaluechartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a single value chart
*/
func getPayloadSingleValueChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("program_text").(string),
	}

	viz := getSingleValueChartOptions(d)
	if vizOptions := getPerSignalVizOptions(d); len(vizOptions) > 0 {
		viz["publishLabelOptions"] = vizOptions
	}
	if len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
}

func getSingleValueChartOptions(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "SingleValue"
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		if val == "Scale" {
			if colorScaleOptions := getColorScaleOptions(d); len(colorScaleOptions) > 0 {
				viz["colorBy"] = "Scale"
				viz["colorScale"] = colorScaleOptions
			}
		} else {
			viz["colorBy"] = val.(string)
		}
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int) * 1000
		viz["programOptions"] = programOptions
	}

	if refreshInterval, ok := d.GetOk("refresh_interval"); ok {
		viz["refreshInterval"] = refreshInterval.(int) * 1000
	}
	if maxPrecision, ok := d.GetOk("max_precision"); ok {
		viz["maximumPrecision"] = maxPrecision.(int)
	}
	viz["timestampHidden"] = d.Get("is_timestamp_hidden").(bool)
	viz["showSparkLine"] = d.Get("show_spark_line").(bool)

	return viz
}

func singlevaluechartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadSingleValueChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.AuthToken, payload, d)
}

func singlevaluechartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.AuthToken, d)
}

func singlevaluechartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadSingleValueChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.AuthToken, payload, d)
}

func singlevaluechartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())
	return resourceDelete(url, config.AuthToken, d)
}
