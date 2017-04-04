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

const DETECTOR_API_URL = "https://api.signalfx.com/v2/detector"

func detectorResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"last_updated": &schema.Schema{
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Latest timestamp the resource was updated",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the detector",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the detector",
			},
			"program_text": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Signalflow program text for the detector. More info at \"https://developers.signalfx.com/docs/signalflow-overview\"",
			},
			"max_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How long (in seconds) to wait for late datapoints",
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
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the rule",
						},
						"notifications": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Determines where notifications will be sent when an incident occurs. See https://developers.signalfx.com/v2/docs/detector-model#notifications-models for more info",
						},
						"severity": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateSeverity,
							Description:  "The severity of the rule, must be one of: Critical, Warning, Major, Minor, Info",
						},
						"detect_label": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A detect label which matches a detect label within the program text",
						},
						"disabled": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "(default: false) When true, notifications and events will not be generated for the detect label",
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
func getPayloadDetector(d *schema.ResourceData) ([]byte, error) {

	tf_rules := d.Get("rule").(*schema.Set).List()
	rules_list := make([]map[string]interface{}, len(tf_rules))
	for i, tf_rule := range tf_rules {
		tf_rule := tf_rule.(map[string]interface{})
		item := make(map[string]interface{})

		item["description"] = tf_rule["description"].(string)
		item["severity"] = tf_rule["severity"].(string)
		item["detectLabel"] = tf_rule["detect_label"].(string)
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
		"programText": d.Get("program_text").(string),
		"maxDelay":    nil,
		"rules":       rules_list,
	}

	if val, ok := d.GetOk("max_delay"); ok {
		payload["maxDelay"] = val.(int) * 1000
	}

	if viz := getVisualizationOptionsDetector(d); len(viz) > 0 {
		payload["visualizationOptions"] = viz
	}

	return json.Marshal(payload)
}

func getVisualizationOptionsDetector(d *schema.ResourceData) map[string]interface{} {
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
			item["credential_id"] = vars[1]
		} else if vars[0] == "Webhook" {
			item["secret"] = vars[1]
			item["url"] = vars[2]
		}

		notifications_list[i] = item
	}

	return notifications_list
}

func detectorCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadDetector(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(DETECTOR_API_URL, config.SfxToken, payload, d)
}

func detectorRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceRead(url, config.SfxToken, d)
}

func detectorUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadDetector(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceUpdate(url, config.SfxToken, payload, d)
}

func detectorDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceDelete(url, config.SfxToken, d)
}

/*
   Hashing function for rule substructure of the detector resource, used in determining state changes.
*/
func resourceRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["description"]))
	buf.WriteString(fmt.Sprintf("%s-", m["severity"]))
	buf.WriteString(fmt.Sprintf("%s-", m["detect_label"]))
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
  Validates the severity field against a list of allowed words.
*/
func validateSeverity(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	allowedWords := []string{"Critical", "Warning", "Major", "Minor", "Info"}
	for _, word := range allowedWords {
		if value == word {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%s not allowed; must be one of: %s", value, strings.Join(allowedWords, ", ")))
	return
}
