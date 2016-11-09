package signalform

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestSignalformConfigureFileNotFound(t *testing.T) {
	// Testing file doesn't exist
	_, err := signalformConfigure(nil)
	assert.Contains(t, err.Error(), "Failed opening config file")
}

func TestSignalformConfigureParseError(t *testing.T) {
	// Testing file exists but json is invalid
	testfile := "config.json"
	// Only run the test when the original config.json file doesn't exist in cwd
	// Provider instantiation is managed by the internal schema framework
	// so the config filename can't be passed in as an argument
	if _, err := os.Stat(testfile); os.IsNotExist(err) {
		_, err := os.Create(testfile)
		if err == nil {
			defer os.Remove(testfile)
			_, err = signalformConfigure(nil)
			assert.Contains(t, err.Error(), "Failed parsing config file")
		}
	}
}

func TestSignalformConfigureSuccess(t *testing.T) {
	// Testing successfully parsed configuration
	testfile := "config.json"
	// Only run the test when the original config.json file doesn't exist in cwd
	if _, err := os.Stat(testfile); os.IsNotExist(err) {
		configFile, err := os.Create(testfile)
		if err == nil {
			configFile.WriteString(`{"detector_endpoint":"test_endpoint","sfx_token":"token"}`)
			defer os.Remove(testfile)
			_, err := signalformConfigure(nil)
			assert.Nil(t, err)
		}
	}
}
