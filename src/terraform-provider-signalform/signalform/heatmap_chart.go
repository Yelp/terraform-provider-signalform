package signalform

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func heatmapChartResource() *schema.Resource {
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
			"minimum_resolution": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum resolution (in seconds) to use for computing the underlying program",
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSortBy,
				Description:  "The property to use when sorting the elements. Must be prepended with + for ascending or - for descending (e.g. -foo)",
			},
			"color_range": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Values and color for the color range. Example: colorRange : { min : 0, max : 100, color : \"blue\" }",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     -math.MaxFloat32,
							Description: "The minimum value within the coloring range",
						},
						"max_value": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "The maximum value within the coloring range",
						},
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The color range to use. Must be either \"gray\", \"blue\", \"navy\", \"orange\", \"yellow\", \"magenta\", \"purple\", \"violet\", \"lilac\", \"green\", \"aquamarine\"",
							ValidateFunc: validateHeatmapChartColor,
						},
					},
				},
			},
			"color_scale": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Single color range including both the color to display for that range and the borders of the range",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gt": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the lower threshold non-inclusive value for this range",
						},
						"gte": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the lower threshold inclusive value for this range",
						},
						"lt": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the upper threshold non-inculsive value for this range",
						},
						"lte": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the upper threshold inclusive value for this range",
						},
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The color to use. Must be either \"gray\", \"blue\", \"navy\", \"orange\", \"yellow\", \"magenta\", \"purple\", \"violet\", \"lilac\", \"green\", \"aquamarine\"",
							ValidateFunc: validateHeatmapChartColor,
						},
					},
				},
			},
			"hide_timestamp": &schema.Schema{
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
		"programText": sanitizeProgramText(d.Get("program_text").(string)),
	}

	viz := getHeatmapOptionsChart(d)
	if len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
}

func getHeatmapColorRangeOptions(d *schema.ResourceData) map[string]interface{} {
	item := make(map[string]interface{})
	colorRange := d.Get("color_range").(*schema.Set).List()
	for _, options := range colorRange {
		options := options.(map[string]interface{})

		if val, ok := options["min_value"]; ok {
			if val.(float64) != -math.MaxFloat32 {
				item["min"] = val.(float64)
			}
		}
		if val, ok := options["max_value"]; ok {
			if val.(float64) != math.MaxFloat32 {
				item["max"] = val.(float64)
			}
		}
		color := options["color"].(string)
		for _, colorStruct := range ChartColorsSlice {
			if color == colorStruct.name {
				item["color"] = colorStruct.name
				break
			}
		}
	}
	return item
}

func getHeatmapOptionsChart(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "Heatmap"
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("minimum_resolution"); ok {
		programOptions["minimumResolution"] = val.(int) * 1000
	}
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int) * 1000
	}
	programOptions["disableSampling"] = d.Get("disable_sampling").(bool)
	viz["programOptions"] = programOptions

	if groupByOptions, ok := d.GetOk("group_by"); ok {
		viz["groupBy"] = groupByOptions.([]interface{})
	}

	if sortProperty, ok := d.GetOk("sort_by"); ok {
		sortBy := sortProperty.(string)
		viz["sortProperty"] = sortBy[1:]
		if strings.HasPrefix(sortBy, "+") {
			viz["sortDirection"] = "Ascending"
		} else {
			viz["sortDirection"] = "Descending"
		}
	}

	if colorRangeOptions := getHeatmapColorRangeOptions(d); len(colorRangeOptions) > 0 {
		viz["colorBy"] = "Range"
		viz["colorRange"] = colorRangeOptions
	} else if colorScaleOptions := getColorScaleOptions(d); len(colorScaleOptions) > 0 {
		viz["colorBy"] = "Scale"
		viz["colorScale2"] = colorScaleOptions
	}

	viz["timestampHidden"] = d.Get("hide_timestamp").(bool)

	return viz
}

func heatmapchartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadHeatmapChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(CHART_API_URL, config.AuthToken, payload, d)
}

func heatmapchartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceRead(url, config.AuthToken, d)
}

func heatmapchartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadHeatmapChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())

	return resourceUpdate(url, config.AuthToken, payload, d)
}

func heatmapchartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", CHART_API_URL, d.Id())
	return resourceDelete(url, config.AuthToken, d)
}

/*
  Validates the color_range field against a list of allowed words.
*/
func validateHeatmapChartColor(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	keys := make([]string, 0, len(ChartColorsSlice))
	found := false
	for _, item := range ChartColorsSlice {
		if value == item.name {
			found = true
		}
		keys = append(keys, item.name)
	}
	if !found {
		joinedColors := strings.Join(keys, ",")
		errors = append(errors, fmt.Errorf("%s not allowed; must be either %s", value, joinedColors))
	}
	return
}
