package signalform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Workaround for Signalfx bug related to post processing and lastUpdatedTime
	OFFSET        = 10000.0
	CHART_API_URL = "https://api.signalfx.com/v2/chart"
)

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
  Validates the time_span_type field against a list of allowed words.
*/
func validateTimeSpanType(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "relative" && value != "absolute" {
		errors = append(errors, fmt.Errorf("%s not allowed; must be either relative or absolute", value))
	}
	return
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
	} else {
		if strings.Contains(string(resp_body), "Resource not found") {
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
	Util method to validate time either in milliseconds since epoch or SignalFx specific string format.
*/
func validateTime(v interface{}, k string) (we []string, errors []error) {
	ts := v.(string)

	// Try to guess if time is in milliseconds since epoch
	ms, err := strconv.Atoi(ts)
	if err == nil {
		if ms < 621129600 {
			errors = append(errors, fmt.Errorf("%s not allowed. Please use milliseconds from epoch or SignalFx time syntax (e.g. -5m, -1h, Now)", ts))
		}
		return
	}

	// If it's the end time, it can only be ms since epoch or now
	if k == "time_end" {
		if ts != "Now" {
			errors = append(errors, fmt.Errorf("%s not allowed. Please use milliseconds from epoch or Now", ts))
		}
		return
	}

	// If it's the start time, it can only be ms since epoch or SignalFx time syntax, but not Now
	r, _ := regexp.Compile("-([0-9]+)[smhdw]")
	if !r.MatchString(ts) {
		errors = append(errors, fmt.Errorf("%s not allowed. Please use milliseconds from epoch or SignalFx time syntax (e.g. -5m, -1h)", ts))
	}
	return
}
