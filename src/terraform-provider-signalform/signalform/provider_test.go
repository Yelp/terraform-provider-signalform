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

func createTempConfigFile(content string, name string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), name)
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

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatal(err.Error())
	}
}

func TestProviderConfigureFromNothing(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "auth_token: required field is not set")
}

func TestProviderConfigureFromTerraform(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Unsetenv("SFX_AUTH_TOKEN")
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

func TestProviderConfigureFromTerraformOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

func TestProviderConfigureFromEnvironment(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Unsetenv("SFX_AUTH_TOKEN")
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "YYY", configuration.AuthToken)
}

func TestProviderConfigureFromEnvironmentOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Unsetenv("SFX_AUTH_TOKEN")
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "YYY", configuration.AuthToken)
}

func TestSignalformConfigureFromHomeFile(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	HomeConfigPath = tmpfileHome.Name()
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalformConfigureFromNetrcFile(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`machine api.signalfx.com login auth_login password WWW`, ".netrc")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	os.Setenv("NETRC", tmpfileHome.Name())
	defer os.Unsetenv("NETRC")
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalformConfigureFromHomeFileOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	HomeConfigPath = tmpfileHome.Name()
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalformConfigureFromSystemFileOnly(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	HomeConfigPath = "filedoesnotexist"
	raw := make(map[string]interface{})
	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("Error creating mock config: %s", err.Error())
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalformConfig)
	assert.Equal(t, "ZZZ", configuration.AuthToken)
}

func TestReadConfigFileFileNotFound(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	config := signalformConfig{}
	err := readConfigFile("foo.conf", &config)
	assert.Contains(t, err.Error(), "Failed to open config file")
}

func TestReadConfigFileParseError(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := createTempConfigFile(`{"auth_tok`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Contains(t, err.Error(), "Failed to parse config file")
}

func TestReadConfigFileSuccess(t *testing.T) {
	config := signalformConfig{}
	tmpfile, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"XXX"}`, "signalform.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Nil(t, err)
	assert.Equal(t, "XXX", config.AuthToken)
}
