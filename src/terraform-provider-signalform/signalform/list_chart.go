package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func listChartResource() *schema.Resource {
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
			"resource_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     CHART_URL,
				Description: "API URL of the chart",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the chart",
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
				Description: "(Metric by default) Must be \"Metric\" or \"Dimension\"",
			},
			"max_delay": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints",
				ValidateFunc: validateMaxDelayValue,
			},
			"disable_sampling": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) If false, samples a subset of the output MTS, which improves UI performance",
			},
			"sort_by": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSortBy,
				Description:  "The property to use when sorting the elements. Use 'value' if you want to sort by value. Must be prepended with + for ascending or - for descending (e.g. -foo)",
			},
			"refresh_interval": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How often (in seconds) to refresh the values of the list",
			},
			"legend_fields_to_hide": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of properties that shouldn't be displayed in the chart legend (i.e. dimension names)",
			},
			"max_precision": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of digits to display when rounding values up or down",
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
						"value_unit": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateUnitTimeChart,
							Description:  "A unit to attach to this plot. Units support automatic scaling (eg thousands of bytes will be displayed as kilobytes)",
						},
						"value_prefix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary prefix to display with the value of this plot",
						},
						"value_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary suffix to display with the value of this plot",
						},
					},
				},
			},
		},

		Create: listchartCreate,
		Read:   listchartRead,
		Update: listchartUpdate,
		Delete: listchartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a list chart
*/
func getPayloadListChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": sanitizeProgramText(d.Get("program_text").(string)),
	}

	viz := getListChartOptions(d)
	if legendOptions := getLegendOptions(d); len(legendOptions) > 0 {
		viz["legendOptions"] = legendOptions
	}
	if vizOptions := getPerSignalVizOptions(d); len(vizOptions) > 0 {
		viz["publishLabelOptions"] = vizOptions
	}
	if len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
}

func getListChartOptions(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "List"
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		viz["colorBy"] = val.(string)
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int) * 1000
	}
	programOptions["disableSampling"] = d.Get("disable_sampling").(bool)
	viz["programOptions"] = programOptions

	if sortBy, ok := d.GetOk("sort_by"); ok {
		viz["sortBy"] = sortBy.(string)
	}
	if refreshInterval, ok := d.GetOk("refresh_interval"); ok {
		viz["refreshInterval"] = refreshInterval.(int) * 1000
	}
	if maxPrecision, ok := d.GetOk("max_precision"); ok {
		viz["maximumPrecision"] = maxPrecision.(int)
	}

	return viz
}

func listchartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadListChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.AuthToken, payload, d)
}

func listchartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.AuthToken, d)
}

func listchartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadListChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.AuthToken, payload, d)
}

func listchartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceDelete(url, config.AuthToken, d)
}
