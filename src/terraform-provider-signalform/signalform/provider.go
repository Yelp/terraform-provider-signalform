package signalform

import (
	"encoding/json"
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"runtime"
)

var SystemConfigPath = "/etc/signalfx.conf"
var HomeConfigSuffix = "/.signalfx.conf"
var HomeConfigPath = ""

type signalformConfig struct {
	AuthToken string `json:"auth_token"`
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SFX_AUTH_TOKEN", nil),
				Description: "SignalFx auth token",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"signalform_detector":           detectorResource(),
			"signalform_time_chart":         timeChartResource(),
			"signalform_heatmap_chart":      heatmapChartResource(),
			"signalform_single_value_chart": singleValueChartResource(),
			"signalform_list_chart":         listChartResource(),
			"signalform_text_chart":         textChartResource(),
			"signalform_dashboard":          dashboardResource(),
			"signalform_dashboard_group":    dashboardGroupResource(),
		},
		ConfigureFunc: signalformConfigure,
	}
}

func signalformConfigure(data *schema.ResourceData) (interface{}, error) {
	config := signalformConfig{}

	// /etc/signalfx.conf has lowest priority
	log.Printf("[DEBUG] Looking for config in system config (%s)...\n", SystemConfigPath)
	if _, err := os.Stat(SystemConfigPath); err == nil {
		log.Printf("[DEBUG] Found %s!\n", SystemConfigPath)
		err = readConfigFile(SystemConfigPath, &config)
		if err != nil {
			log.Printf("[DEBUG] Failed reading system config: %s\n", err.Error())
			return nil, err
		} else {
			log.Printf("[DEBUG] Parsed system config.")
		}
	} else {
		log.Printf("[DEBUG] Could not find %s\n", SystemConfigPath)
	}

	// $HOME/.signalfx.conf second
	// this additional variable is used for mocking purposes in tests
	if HomeConfigPath == "" {
		usr, err := user.Current()
		HomeConfigPath = usr.HomeDir + HomeConfigSuffix
		if err != nil {
			return nil, fmt.Errorf("Failed to get user environment %s", err.Error())
		}
	}

	log.Printf("[DEBUG] Looking for config in home dir (%s)\n", HomeConfigPath)
	if _, err := os.Stat(HomeConfigPath); err == nil {
		log.Printf("[DEBUG] Found %s\n", HomeConfigPath)
		err = readConfigFile(HomeConfigPath, &config)
		if err != nil {
			log.Printf("[DEBUG] Failed reading home dir config: %s\n", err.Error())
			return nil, err
		} else {
			log.Printf("[DEBUG] Parsed home dir config.")
		}
	} else {
		log.Printf("[DEBUG] Could not find %s\n", HomeConfigPath)
	}

	// Use netrc next
	err := readNetrcFile(&config)
	if err != nil {
		return nil, err
	}

	// provider is the top priority
	if token, ok := data.GetOk("auth_token"); ok {
		log.Printf("[DEBUG] Reading config from provider.\n")
		log.Printf("[DEBUG] config.AuthToken has length %d provider data auth token has length %d", len(config.AuthToken), len(token.(string)))
		config.AuthToken = token.(string)
	} else {
		log.Printf("[DEBUG] Did not find config in provider.\n")
	}

	if len(config.AuthToken) == 0 {
		log.Printf("[DEBUG] config.AuthToken has length %d", len(config.AuthToken))
		return &config, fmt.Errorf("auth_token: required field is not set")
	} else {
		log.Printf("[DEBUG] config.AuthToken is longer than 0 bytes")
	}

	return &config, nil
}

func readConfigFile(configPath string, config *signalformConfig) error {
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Failed to open config file. %s", err.Error())
	}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return fmt.Errorf("Failed to parse config file (%s): %s", configPath, err.Error())
	}
	return nil
}

func readNetrcFile(config *signalformConfig) error {
	// Inspired by https://github.com/hashicorp/terraform/blob/master/vendor/github.com/hashicorp/go-getter/netrc.go
	// Get the netrc file path
	path := os.Getenv("NETRC")
	if path == "" {
		filename := ".netrc"
		if runtime.GOOS == "windows" {
			filename = "_netrc"
		}

		var err error
		path, err = homedir.Expand("~/" + filename)
		if err != nil {
			return err
		}
	}

	// If the file is not a file, then do nothing
	if fi, err := os.Stat(path); err != nil {
		// File doesn't exist, do nothing
		if os.IsNotExist(err) {
			return nil
		}

		// Some other error!
		return err
	} else if fi.IsDir() {
		// File is directory, ignore
		return nil
	}

	// Load up the netrc file
	net, err := netrc.ParseFile(path)
	if err != nil {
		return fmt.Errorf("Error parsing netrc file at %q: %s", path, err)
	}

	machine := net.FindMachine("api.signalfx.com")
	if machine == nil {
		// Machine not found, no problem
		return nil
	}

	// Set the auth token
	config.AuthToken = machine.Password
	return nil
}
