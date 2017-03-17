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
		payload["visualizationOptions"] = viz
	}

	return json.Marshal(payload)
}

func getVisualizationOptionsChart(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
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
