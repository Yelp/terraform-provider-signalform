package signalform

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err.Error())
	}
}

func TestProviderConfigureEmptyConfig(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigSuffix = "/.filedoesnotexist"
	rp := Provider()
	raw := map[string]interface{}{}

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "auth_token: required field is not set")

}

func TestProviderConfigureFromTerraform(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigSuffix = "/.filedoesnotexist"
	rp := Provider()
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}

	configuration := meta.(*signalformConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

func TestProviderConfigureFromEnvironment(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigSuffix = "/.filedoesnotexist"
	rp := Provider()
	raw := make(map[string]interface{})
	os.Setenv("SFX_AUTH_TOKEN", "XXX")
	defer os.Unsetenv("SFX_AUTH_TOKEN")

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}

	configuration := meta.(*signalformConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

/*
TODO: Upgrade to terraform 0.9 so we can use TestResourceDataRaw and do the testing for real
func TestSignalformConfigureFromData(t *testing.T) {
	rp := Provider().(*schema.Provider)
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}
	data := TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromHomeFile(t *testing.T) {
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	// mock reading home file somehow
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromSystemFile(t *testing.T) {
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	// mock reading home file somehow
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}
*/

func TestSignalformConfigureFileNotFound(t *testing.T) {
	config := signalformConfig{}
	err := readConfigFile("foo.conf", &config)
	assert.Contains(t, err.Error(), "Failed to open config file")
}

func TestSignalformConfigureParseError(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := ioutil.TempFile(os.TempDir(), "signalform")
	if err != nil {
		t.Fatalf("Error creating temporary test file. err: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Contains(t, err.Error(), "Failed to parse config file")
}

func TestSignalformConfigureSuccess(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := ioutil.TempFile(os.TempDir(), "signalform")
	if err != nil {
		t.Fatalf("Error creating temporary test file. err: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`{"useless_config":"foo","auth_token":"XXX"}`)
	if err != nil {
		t.Fatalf("Error writing to temporary test file. err: %s", err.Error())
	}

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Nil(t, err)
	assert.Equal(t, "XXX", config.AuthToken)
}
