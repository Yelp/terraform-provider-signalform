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

func TestProviderConfigureFromTerraform(t *testing.T) {
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
	assert.Equal(t, "XXX", configuration.SfxToken)
}

func TestProviderConfigureFromEnvironment(t *testing.T) {
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
	assert.Equal(t, "XXX", configuration.SfxToken)
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
	assert.Equal(t, "XXX", configuration.(*signalformConfig).SfxToken)
}

func TestSignalformConfigureFromHomeFile(t *testing.T) {
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	// mock reading home file somehow
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).SfxToken)
}

func TestSignalformConfigureFromSystemFile(t *testing.T) {
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	// mock reading home file somehow
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).SfxToken)
}
*/

func TestSignalformConfigureFileNotFound(t *testing.T) {
	_, err := readConfigFile("foo.conf")
	assert.Contains(t, err.Error(), "Failed opening config file")
}

func TestSignalformConfigureParseError(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "signalform")
	if err != nil {
		t.Fatalf("Error creating temporary test file. err: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = readConfigFile(tmpfile.Name())
	assert.Contains(t, err.Error(), "Failed parsing config file")
}

func TestSignalformConfigureSuccess(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "signalform")
	if err != nil {
		t.Fatalf("Error creating temporary test file. err: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`{"useless_config":"foo","auth_token":"XXX"}`)
	if err != nil {
		t.Fatalf("Error writing to temporary test file. err: %s", err.Error())
	}

	meta, err := readConfigFile(tmpfile.Name())
	assert.Nil(t, err)
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "XXX", configuration.SfxToken)
}
