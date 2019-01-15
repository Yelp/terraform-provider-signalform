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

const (
	DETECTOR_API_URL = "https://api.signalfx.com/v2/detector"
	DETECTOR_URL     = "https://app.signalfx.com/#/detector/v2/<id>/edit"
)

func detectorResource() *schema.Resource {
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
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Url of the detector",
			},
			"resource_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     DETECTOR_URL,
				Description: "Base Detector url",
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
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints. Max value 900s (15m)",
				ValidateFunc: validateMaxDelayValue,
			},
			"show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) When true, markers will be drawn for each datapoint within the visualization.",
			},
			"time_range": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validateSignalfxRelativeTime,
				Description:   "From when to display data. SignalFx time syntax (e.g. -5m, -1h)",
				ConflictsWith: []string{"start_time", "end_time"},
			},
			"start_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Seconds since epoch. Used for visualization",
			},
			"end_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Seconds since epoch. Used for visualization",
			},
			"tags": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the detector",
			},
			"teams": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Team IDs to associate the detector to",
			},
			"rule": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of rules used for alerting",
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
							Description: "List of strings specifying where notifications will be sent when an incident occurs. See https://developers.signalfx.com/v2/docs/detector-model#notifications-models for more info",
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
						"parameterized_body": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom notification message body when an alert is triggered. See https://developers.signalfx.com/v2/reference#detector-model for more info",
						},
						"parameterized_subject": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom notification message subject when an alert is triggered. See https://d    evelopers.signalfx.com/v2/reference#detector-model for more info",
						},
						"runbook_url": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of page to consult when an alert is triggered",
						},
						"tip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Plain text suggested first course of action, such as a command to execute.",
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

		if val, ok := tf_rule["parameterized_body"]; ok {
			item["parameterizedBody"] = val.(string)
		}

		if val, ok := tf_rule["parameterized_subject"]; ok {
			item["parameterizedSubject"] = val.(string)
		}

		if val, ok := tf_rule["runbook_url"]; ok {
			item["runbookUrl"] = val.(string)
		}

		if val, ok := tf_rule["tip"]; ok {
			item["tip"] = val.(string)
		}

		if notifications, ok := tf_rule["notifications"]; ok {
			notify := getNotifications(notifications.([]interface{}))
			item["notifications"] = notify
		}

		rules_list[i] = item
	}

	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": sanitizeProgramText(d.Get("program_text").(string)),
		"maxDelay":    nil,
		"rules":       rules_list,
	}

	if val, ok := d.GetOk("max_delay"); ok {
		payload["maxDelay"] = val.(int) * 1000
	}

	if viz := getVisualizationOptionsDetector(d); len(viz) > 0 {
		payload["visualizationOptions"] = viz
	}

	if val, ok := d.GetOk("teams"); ok {
		teams := []string{}
		for _, team := range val.([]interface{}) {
			teams = append(teams, team.(string))
		}
		payload["teams"] = teams
	}

	if val, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, tag := range val.([]interface{}) {
			tags = append(tags, tag.(string))
		}
		payload["tags"] = tags
	}

	return json.Marshal(payload)
}

func getVisualizationOptionsDetector(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	if val, ok := d.GetOk("show_data_markers"); ok {
		viz["showDataMarkers"] = val.(bool)
	}

	timeMap := make(map[string]interface{})
	if val, ok := d.GetOk("time_range"); ok {
		if ms, err := fromRangeToMilliSeconds(val.(string)); err == nil {
			timeMap["range"] = ms
			timeMap["type"] = "relative"
		}
	}
	if val, ok := d.GetOk("start_time"); ok {
		timeMap["type"] = "absolute"
		timeMap["start"] = val.(int) * 1000
		if val, ok := d.GetOk("end_time"); ok {
			timeMap["end"] = val.(int) * 1000
		}
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
		} else if vars[0] == "Slack" {
			item["credentialId"] = vars[1]
			item["channel"] = vars[2]
		} else if vars[0] == "Webhook" {
			item["secret"] = vars[1]
			item["url"] = vars[2]
		} else if vars[0] == "Team" || vars[0] == "TeamEmail" {
			item["team"] = vars[1]
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

	return resourceCreate(DETECTOR_API_URL, config.AuthToken, payload, d)
}

func detectorRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceRead(url, config.AuthToken, d)
}

func detectorUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadDetector(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceUpdate(url, config.AuthToken, payload, d)
}

func detectorDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DETECTOR_API_URL, d.Id())

	return resourceDelete(url, config.AuthToken, d)
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

	// loop through optional rule attributes
	var optional_rule_keys = []string{"parameterized_body", "parameterized_subject", "runbook_url", "tip"}

	for _, key := range optional_rule_keys {
		if val, ok := m[key]; ok {
			buf.WriteString(fmt.Sprintf("%s-", val))
		}
	}

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
	allowedWords := []string{"Critical", "Major", "Minor", "Warning", "Info"}
	for _, word := range allowedWords {
		if value == word {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%s not allowed; must be one of: %s", value, strings.Join(allowedWords, ", ")))
	return
}
