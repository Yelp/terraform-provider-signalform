package signalform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

const (
	// Workaround for Signalfx bug related to post processing and lastUpdatedTime
	OFFSET        = 10000.0
	CHART_API_URL = "https://api.signalfx.com/v2/chart"
	CHART_URL     = "https://app.signalfx.com/#/chart/<id>"
)

type chartColor struct {
	name string
	hex  string
}

var ChartColorsSlice = []chartColor{
	{"gray", "#999999"},
	{"blue", "#0077c2"},
	{"light_blue", "#00b9ff"},
	{"navy", "#6CA2B7"},
	{"dark_orange", "#b04600"},
	{"orange", "#f47e00"},
	{"dark_yellow", "#e5b312"},
	{"magenta", "#bd468d"},
	{"cerise", "#e9008a"},
	{"pink", "#ff8dd1"},
	{"violet", "#876ff3"},
	{"purple", "#a747ff"},
	{"gray_blue", "#ab99bc"},
	{"dark_green", "#007c1d"},
	{"green", "#05ce00"},
	{"aquamarine", "#0dba8f"},
	{"red", "#ea1849"},
	{"yellow", "#ea1849"},
	{"vivid_yellow", "#ea1849"},
	{"light_green", "#acef7f"},
	{"lime_green", "#6bd37e"},
}

/*
  Utility function that wraps http calls to SignalFx
*/
func sendRequest(method string, url string, token string, payload []byte) (int, []byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-SF-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return -1, nil, fmt.Errorf("Failed sending %s request to Signalfx: %s", method, err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("Failed reading response body from %s request: %s", method, err.Error())
	}

	return resp.StatusCode, body, nil
}

/*
  Validates max_delay field; it must be between 0 and 900 seconds (15m in).
*/
func validateMaxDelayValue(v interface{}, k string) (we []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 900 {
		errors = append(errors, fmt.Errorf("%d not allowed; max_delay must be >= 0 && <= 900", value))
	}
	return
}

/*
  Validates that sort_by field start with either + or -.
*/
func validateSortBy(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if !strings.HasPrefix(value, "+") && !strings.HasPrefix(value, "-") {
		errors = append(errors, fmt.Errorf("%s not allowed; must start either with + or - (ascending or descending)", value))
	}
	return
}

/*
	Get Color Scale Options
*/
func getColorScaleOptions(d *schema.ResourceData) []interface{} {
	colorScale := d.Get("color_scale").(*schema.Set).List()
	return getColorScaleOptionsFromSlice(colorScale)
}

func getColorScaleOptionsFromSlice(colorScale []interface{}) []interface{} {
	item := make([]interface{}, len(colorScale))
	if len(colorScale) == 0 {
		return item
	}
	for i := range colorScale {
		options := make(map[string]interface{})
		scale := colorScale[i].(map[string]interface{})
		if scale["gt"].(float64) != math.MaxFloat32 {
			options["gt"] = scale["gt"].(float64)
		}
		if scale["gte"].(float64) != math.MaxFloat32 {
			options["gte"] = scale["gte"].(float64)
		}
		if scale["lt"].(float64) != math.MaxFloat32 {
			options["lt"] = scale["lt"].(float64)
		}
		if scale["lte"].(float64) != math.MaxFloat32 {
			options["lte"] = scale["lte"].(float64)
		}
		paletteIndex := 0
		for index, thing := range ChartColorsSlice {
			if scale["color"] == thing.name {
				paletteIndex = index
				break
			}
		}
		options["paletteIndex"] = paletteIndex
		item[i] = options
	}
	return item
}

/*
  Send a GET to get the current state of the resource. It just checks if the lastUpdated timestamp is
  later than the timestamp saved in the resource. If so, the resource has been modified in some way
  in the UI, and should be recreated. This is signaled by setting synced to false, meaning if synced is set to
  true in the tf configuration, it will update the resource to achieve the desired state.
*/
func resourceRead(url string, sfxToken string, d *schema.ResourceData) error {
	status_code, resp_body, err := sendRequest("GET", url, sfxToken, nil)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for the resource %s during read: %s", d.Get("name"), err.Error())
		}
		// This implies the resource was modified in the Signalfx UI and therefore it is not synced with Signalform
		last_updated := mapped_resp["lastUpdated"].(float64)
		if last_updated > (d.Get("last_updated").(float64) + OFFSET) {
			d.Set("synced", false)
			d.Set("last_updated", last_updated)
		}
		var resource_url string
		if val, ok := d.GetOk("resource_url"); ok {
			resource_url = strings.Replace(fmt.Sprintf("%s", val), "<id>", mapped_resp["id"].(string), 1)
		} else {
			resource_url = "DUMMY"
		}
		d.Set("url", resource_url)
	} else {
		if status_code == 404 && strings.Contains(string(resp_body), " not found") {
			// This implies that the resouce was deleted in the Signalfx UI and therefore we need to recreate it
			d.SetId("")
		} else {
			return fmt.Errorf("For the resource %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
		}
	}

	return nil
}

/*
  Fetches payload specified in terraform configuration and creates a resource
*/
func resourceCreate(url string, sfxToken string, payload []byte, d *schema.ResourceData) error {
	status_code, resp_body, err := sendRequest("POST", url, sfxToken, payload)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for the resource %s during creation: %s", d.Get("name"), err.Error())
		}
		d.SetId(fmt.Sprintf("%s", mapped_resp["id"].(string)))
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
		d.Set("synced", true)
		// Replace "<id>" with the actual Resource ID
		resource_url := strings.Replace(fmt.Sprintf("%s", d.Get("resource_url")), "<id>", mapped_resp["id"].(string), 1)
		d.Set("url", resource_url)
	} else {
		return fmt.Errorf("For the resource %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Fetches payload specified in terraform configuration and creates chart
*/
func resourceUpdate(url string, sfxToken string, payload []byte, d *schema.ResourceData) error {
	status_code, resp_body, err := sendRequest("PUT", url, sfxToken, payload)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for the resource %s during creation: %s", d.Get("name"), err.Error())
		}
		// If the resource was updated successfully with Signalform configs, it is now synced with Signalfx
		d.Set("synced", true)
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
		resource_url := strings.Replace(fmt.Sprintf("%s", d.Get("resource_url")), "<id>", mapped_resp["id"].(string), 1)
		d.Set("url", resource_url)
	} else {
		return fmt.Errorf("For the resource %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Deletes a resource.  If the resource does not exist, it will receive a 404, and carry on as usual.
*/
func resourceDelete(url string, sfxToken string, d *schema.ResourceData) error {
	status_code, resp_body, err := sendRequest("DELETE", url, sfxToken, nil)
	if err != nil {
		return fmt.Errorf("Failed deleting resource  %s: %s", d.Get("name"), err.Error())
	}
	if status_code < 400 || status_code == 404 {
		d.SetId("")
	} else {
		return fmt.Errorf("For the resource  %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
	Util method to get Legend Chart Options.
*/
func getLegendOptions(d *schema.ResourceData) map[string]interface{} {
	if properties, ok := d.GetOk("legend_fields_to_hide"); ok {
		properties := properties.(*schema.Set).List()
		legendOptions := make(map[string]interface{})
		properties_opts := make([]map[string]interface{}, len(properties))
		for i, property := range properties {
			property := property.(string)
			if property == "metric" {
				property = "sf_originatingMetric"
			} else if property == "plot_label" || property == "Plot Label" {
				property = "sf_metric"
			}
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

/*
	Util method to validate SignalFx specific string format.
*/
func validateSignalfxRelativeTime(v interface{}, k string) (we []string, errors []error) {
	ts := v.(string)

	r, _ := regexp.Compile("-([0-9]+)[mhdw]")
	if !r.MatchString(ts) {
		errors = append(errors, fmt.Errorf("%s not allowed. Please use milliseconds from epoch or SignalFx time syntax (e.g. -5m, -1h)", ts))
	}
	return
}

/*
*  Util method to convert from Signalfx string format to milliseconds
 */
func fromRangeToMilliSeconds(timeRange string) (int, error) {
	r := regexp.MustCompile("-([0-9]+)([mhdw])")
	ss := r.FindStringSubmatch(timeRange)
	var c int
	switch ss[2] {
	case "m":
		c = 60 * 1000
	case "h":
		c = 60 * 60 * 1000
	case "d":
		c = 24 * 60 * 60 * 1000
	case "w":
		c = 7 * 24 * 60 * 60 * 1000
	default:
		c = 1
	}
	val, err := strconv.Atoi(ss[1])
	if err != nil {
		return -1, err
	}
	return val * c, nil
}

/*
  Validates the color field against a list of allowed words.
*/
func validatePerSignalColor(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if _, ok := PaletteColors[value]; !ok {
		keys := make([]string, 0, len(PaletteColors))
		for k := range PaletteColors {
			keys = append(keys, k)
		}
		joinedColors := strings.Join(keys, ",")
		errors = append(errors, fmt.Errorf("%s not allowed; must be either %s", value, joinedColors))
	}
	return
}

/*
  Sanitize program_text to reduce the errors we get back from SignalFx
*/
func sanitizeProgramText(text string) string {
	r, _ := regexp.Compile("\n[\t\n\v\f\r ]+")
	sane := r.ReplaceAllString(text, "\n")
	r, _ = regexp.Compile("^[\t\n\v\f\r ]+")
	sane = r.ReplaceAllString(sane, "")
	return sane
}
