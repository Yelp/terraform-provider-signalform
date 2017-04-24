package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"math"
	"strings"
)

var ChartColors = map[string]string{
	"gray":       "#999999",
	"blue":       "#0077c2",
	"navy":       "#6CA2B7",
	"orange":     "#b04600",
	"yellow":     "#e5b312",
	"magenta":    "#bd468d",
	"purple":     "#e9008a",
	"violet":     "#876ffe",
	"lilac":      "#a747ff",
	"green":      "#05ce00",
	"aquamarine": "#0dba8f",
}

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
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How long (in seconds) to wait for late datapoints",
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
							ValidateFunc: validateChartColor,
						},
					},
				},
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
		"programText": d.Get("program_text").(string),
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
		item["color"] = ChartColors[color]
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
		viz["colorScale"] = colorScaleOptions
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

/*
  Validates the color_range field against a list of allowed words.
*/
func validateChartColor(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if _, ok := ChartColors[value]; !ok {
		errors = append(errors, fmt.Errorf("%s not allowed; must be either gray, blue, navy, orange, yellow, magenta, purple, violet, lilac, green, aquamarine", value))
	}
	return
}
