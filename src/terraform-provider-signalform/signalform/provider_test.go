package signalform

import (
	"fmt"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var OldSystemConfigPath = SystemConfigPath
var OldHomeConfigPath = HomeConfigPath

func resetGlobals() {
	SystemConfigPath = OldSystemConfigPath
	HomeConfigPath = OldHomeConfigPath
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err.Error())
	}
}

func TestProviderConfigureEmptyConfig(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	rp := Provider()
	raw := make(map[string]interface{})

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "auth_token: required field is not set")
}

func TestProviderConfigureFromTerraform(t *testing.T) {
	tmpfile, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"XXX"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())
	SystemConfigPath = tmpfile.Name()
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
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
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
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

func TestSignalformConfigureFromSystemFile(t *testing.T) {
	tmpfile, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"XXX"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())
	SystemConfigPath = tmpfile.Name()
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := schema.TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromHomeFile(t *testing.T) {
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"YYY"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"XXX"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	SystemConfigPath = tmpfileSystem.Name()
	HomeConfigPath = tmpfileHome.Name()
	defer resetGlobals()
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := schema.TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromData(t *testing.T) {
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"YYY"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"XXX"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	SystemConfigPath = tmpfileSystem.Name()
	HomeConfigPath = tmpfileHome.Name()
	defer resetGlobals()
	rp := Provider().(*schema.Provider)
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}
	data := schema.TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromDataNoFiles(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	rp := Provider().(*schema.Provider)
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}
	data := schema.TestResourceDataRaw(t, rp.Schema, raw)
	configuration, err := signalformConfigure(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, "XXX", configuration.(*signalformConfig).AuthToken)
}

func TestSignalformConfigureFromNothing(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	rp := Provider().(*schema.Provider)
	raw := make(map[string]interface{})
	data := schema.TestResourceDataRaw(t, rp.Schema, raw)
	_, err := signalformConfigure(data)
	assert.Contains(t, err.Error(), "auth_token: required field is not set")
}

func TestSignalformConfigureFileNotFound(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	config := signalformConfig{}
	err := readConfigFile("foo.conf", &config)
	assert.Contains(t, err.Error(), "Failed to open config file")
}

func TestSignalformConfigureParseError(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := createTempConfigFile(`{"auth_tok`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Contains(t, err.Error(), "Failed to parse config file")
}

func TestSignalformConfigureSuccess(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"XXX"}`)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Nil(t, err)
	assert.Equal(t, "XXX", config.AuthToken)
}

func createTempConfigFile(content string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "signalform.conf")
	if err != nil {
		return nil, fmt.Errorf("Error creating temporary test file. err: %s", err.Error())
	}

	_, err = tmpfile.WriteString(content)
	if err != nil {
		os.Remove(tmpfile.Name())
		return nil, fmt.Errorf("Error writing to temporary test file. err: %s", err.Error())
	}

	return tmpfile, nil
}
