package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
)

const DASHBOARD_GROUP_API_URL = "https://api.signalfx.com/v2/dashboardgroup"

func dashboardgroupResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"synced": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Setting synced to 1 implies that the dashboard in SignalForm and SignalFx are identical",
			},
			"last_updated": &schema.Schema{
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Latest timestamp the resource was updated",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the dashboard group",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the dashboard group",
			},
		},

		Create: dashboardgroupCreate,
		Read:   dashboardgroupRead,
		Update: dashboardgroupUpdate,
		Delete: dashboardgroupDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a dasboard group
*/
func getPayloadDashboardGroup(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"dashboards":  make([]string, 0),
	}

	a, e := json.Marshal(payload)
	_ = ioutil.WriteFile("/tmp/fdc_chartCreate", a, 0644)
	return a, e
}

func dashboardgroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadDashboardGroup(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	return resourceCreate(DASHBOARD_GROUP_API_URL, config.SfxToken, payload, d)
}

func dashboardgroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DASHBOARD_GROUP_API_URL, d.Id())

	return resourceRead(url, config.SfxToken, d)
}

func dashboardgroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	payload, err := getPayloadDashboardGroup(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url := fmt.Sprintf("%s/%s", DASHBOARD_GROUP_API_URL, d.Id())

	return resourceUpdate(url, config.SfxToken, payload, d)
}

func dashboardgroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalformConfig)
	url := fmt.Sprintf("%s/%s", DASHBOARD_GROUP_API_URL, d.Id())
	return resourceDelete(url, config.SfxToken, d)
}
