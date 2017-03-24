package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"math"
)

func heatmapchartResource() *schema.Resource {
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
			//"color_by": &schema.Schema{
			//	Type:        schema.TypeString,
			//	Optional:    true,
			///	Description: "(Range by default) Must be \"Range\" or \"Scale\". Range maps to Auto and Scale maps to Fixed in the UI",
			//},
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
				Default:     false,
				Description: "(false by default) If false, samples a subset of the output MTS, which improves UI performance",
			},
			"group_by": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Properties to group by in the heatmap (in nesting order)",
			},
			"sort_by": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The property to use when sorting the elements",
			},
			"is_ascending_sorted": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "(true by default) \"Ascending\" or \"Descending\" sorting",
			},
			"color_range": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Values and color for the color range. Example: colorRange : { min : 0, max : 100, color : \"#00FF00\" }",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "The minimum value within the coloring range",
						},
						"max_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "The maximum value within the coloring range",
						},
						"color": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The color range to use",
						},
						"scale_thresholds": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeFloat},
							Description: "The thresholds to set for the color range being used. Values must be in descending order",
						},
						"scale_inverted": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "(false by default) If false or omitted, values are red if they are above the highest specified value. If true, values are red if they are below the lowest specified value",
						},
					},
				},
			},
			"is_timestamp_hidden": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Whether to show the timestamp in the chart",
			},
		},

		Create: heatmapchartCreate,
		Read:   heatmapchartRead,
		Update: heatmapchartUpdate,
		Delete: heatmapchartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create an Heatmap chart
*/
func getPayloadHeatmapChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("program_text").(string),
	}

	viz := getHeatmapOptionsChart(d)
	if len(viz) > 0 {
		payload["options"] = viz
	}

	a, e := json.Marshal(payload)
	_ = ioutil.WriteFile("/tmp/fdc_chartCreate", a, 0644)
	return a, e
}

func getHeatmapOptionsChart(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "Heatmap"
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		viz["colorBy"] = val.(string)
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("minimum_resolution"); ok {
		programOptions["minimumResolution"] = val.(int)
	}
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int)
	}
	programOptions["disableSampling"] = d.Get("disable_sampling").(bool)
	viz["programOptions"] = programOptions

	if groupByOptions, ok := d.GetOk("group_by"); ok {
		viz["groupBy"] = groupByOptions.([]interface{})
	}

	if sortProperty, ok := d.GetOk("sort_by"); ok {
		viz["sortProperty"] = sortProperty.(string)
	}

	if d.Get("is_ascending_sorted").(bool) {
		viz["sortDirection"] = "Ascending"
	} else {
		viz["sortDirection"] = "Descending"
	}

	viz["timestampHidden"] = d.Get("is_timestamp_hidden").(bool)

	return viz
}

func heatmapchartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadHeatmapChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.SfxToken, payload, d)
}

func heatmapchartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.SfxToken, d)
}

func heatmapchartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadHeatmapChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.SfxToken, payload, d)
}

func heatmapchartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())
	return resourceDelete(url, config.SfxToken, d)
}
