package signalform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"sort"
	"strings"
)

func detectorResource() *schema.Resource {
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
			"maxDelay": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) When true, markers will be drawn for each datapoint within the visualization.",
			},
			"time_span_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateTimeSpanType,
				Description:  "The type of time span defined for visualization. Must be either \"relative\" or \"absoulte\".",
			},
			"time_range": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"start_time", "end_time"},
				Description:   "The time range prior to now to visualize, in milliseconds. You must specify time_span_type = \"relative\" too.",
			},
			"start_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Milliseconds since epoch. Used for visualization. You must specify time_span_type = \"absolute\" too.",
			},
			"end_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Milliseconds since epoch. Used for visualization. You must specify time_span_type = \"absolute\" too.",
			},
			"rule": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"notifications": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"severity": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"detectLabel": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"disabled": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				Set: resourceRuleHash,
			},
		},

		Create: detectorCreate,
		Read:   detectorRead,
		Update: detectorUpdate,
		Delete: detectorDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a detector
*/
func getPayload(d *schema.ResourceData) ([]byte, error) {

	tf_rules := d.Get("rule").(*schema.Set).List()
	rules_list := make([]map[string]interface{}, len(tf_rules))
	for i, tf_rule := range tf_rules {
		tf_rule := tf_rule.(map[string]interface{})
		item := make(map[string]interface{})

		item["description"] = tf_rule["description"].(string)
		item["severity"] = tf_rule["severity"].(string)
		item["detectLabel"] = tf_rule["detectLabel"].(string)
		item["disabled"] = tf_rule["disabled"].(bool)

		if notifications, ok := tf_rule["notifications"]; ok {
			notify := getNotifications(notifications.([]interface{}))
			item["notifications"] = notify
		}

		rules_list[i] = item
	}

	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("programText").(string),
		"maxDelay":    nil,
		"rules":       rules_list,
	}

	if _, ok := d.GetOk("maxDelay"); ok {
		payload["maxDelay"] = int(d.Get("maxDelay").(float64))
	}

	if viz := getVisualizationOptions(d); len(viz) > 0 {
		payload["visualizationOptions"] = viz
	}

	return json.Marshal(payload)
}

func getVisualizationOptions(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	if val, ok := d.GetOk("show_data_markers"); ok {
		viz["showDataMarkers"] = val.(bool)
	}

	timeMap := make(map[string]interface{})
	if val, ok := d.GetOk("time_span_type"); ok {
		timeMap["type"] = val.(string)
	}
	if val, ok := d.GetOk("time_range"); ok {
		timeMap["range"] = val.(int)
	}
	if val, ok := d.GetOk("start_time"); ok {
		timeMap["start"] = val.(int)
	}
	if val, ok := d.GetOk("end_time"); ok {
		timeMap["end"] = val.(int)
	}
	if len(timeMap) > 0 {
		viz["time"] = timeMap
	}
	return viz
}

/*
  Get list of notifications from Resource object (a list of strings), and return a list of notification maps
*/
func getNotifications(tf_notifications []interface{}) []map[string]interface{} {
	notifications_list := make([]map[string]interface{}, len(tf_notifications))
	for i, tf_notification := range tf_notifications {
		vars := strings.Split(tf_notification.(string), ",")
		item := make(map[string]interface{})
		item["type"] = vars[0]

		if vars[0] == "Email" {
			item["email"] = vars[1]
		} else if vars[0] == "PagerDuty" {
			item["credentialId"] = vars[1]
		} else if vars[0] == "Webhook" {
			item["secret"] = vars[1]
			item["url"] = vars[2]
		}

		notifications_list[i] = item
	}

	return notifications_list
}

/*
  Fetches payload specified in terraform configuration and creates detector
*/
func detectorCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := config.ProviderEndpoint
	payload, err := getPayload(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	status_code, resp_body, err := sendRequest("POST", url, config.SfxToken, payload)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for detector %s during creation: %s", d.Get("name"), err.Error())
		}
		d.SetId(fmt.Sprintf("%s", mapped_resp["id"].(string)))
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
	} else {
		return fmt.Errorf("For Detector %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Send a GET to get the current state of the detector.  It just checks if the lastUpdated timestamp is
  later than the timestamp saved in the resource.  If so, the detector has been modified in some way
  in the UI, and should be recreated.  This is signaled by setting synced to 0, meaning if synced is set to
  1 in the tf configuration, it will update the detector to achieve the desired state.
*/
func detectorRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())

	status_code, resp_body, err := sendRequest("GET", url, config.SfxToken, nil)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for detector %s during read: %s", d.Get("name"), err.Error())
		}
		// This implies the detector was modified in the Signalfx UI and therefore it is not synced with Signalform
		last_updated := mapped_resp["lastUpdated"].(float64)
		if last_updated > (d.Get("last_updated").(float64) + OFFSET) {
			d.Set("synced", 0)
			d.Set("last_updated", last_updated)
		}
	} else {
		if strings.Contains(string(resp_body), "Detector not found") {
			// This implies detector was deleted in the Signalfx UI and therefore we need to recreate it
			d.SetId("")
		} else {
			return fmt.Errorf("For Detector %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
		}
	}
	return nil
}

func detectorUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayload(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())

	status_code, resp_body, err := sendRequest("PUT", url, config.SfxToken, payload)
	if status_code == 200 {
		mapped_resp := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &mapped_resp)
		if err != nil {
			return fmt.Errorf("Failed unmarshaling for detector %s during creation: %s", d.Get("name"), err.Error())
		}
		// If the detector was updated successfully with Signalform configs, it is now synced with Signalfx
		d.Set("synced", 1)
		d.Set("last_updated", mapped_resp["lastUpdated"].(float64))
	} else {
		return fmt.Errorf("For Detector %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
  Deletes a detector.  If the detector does not exist, it will receive a 404, and carry on as usual.
*/
func detectorDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", config.ProviderEndpoint, d.Id())
	status_code, resp_body, err := sendRequest("DELETE", url, config.SfxToken, nil)
	if err != nil {
		return fmt.Errorf("Failed deleting detector %s: %s", d.Get("name"), err.Error())
	}
	if status_code < 400 || status_code == 404 {
		d.SetId("")
	} else {
		return fmt.Errorf("For Detector %s SignalFx returned status %d: \n%s", d.Get("name"), status_code, resp_body)
	}
	return nil
}

/*
   Hashing function for rule substructure of the detector resource, used in determining state changes.
*/
func resourceRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["description"]))
	buf.WriteString(fmt.Sprintf("%s-", m["severity"]))
	buf.WriteString(fmt.Sprintf("%s-", m["detectLabel"]))
	buf.WriteString(fmt.Sprintf("%s-", m["disabled"]))

	// Sort the notifications so that we generate a consistent hash
	if v, ok := m["notifications"]; ok {
		notifications := v.([]interface{})
		s_notifications := make([]string, len(notifications))
		for i, raw := range notifications {
			s_notifications[i] = raw.(string)
		}
		sort.Strings(s_notifications)

		for _, notification := range s_notifications {
			buf.WriteString(fmt.Sprintf("%s-", notification))
		}
	}

	return hashcode.String(buf.String())
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
