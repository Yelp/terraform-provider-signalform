package signalform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateChartsResolutionAllowed(t *testing.T) {
	for _, value := range []string{"default", "low", "high", "highest"} {
		_, errors := validateChartsResolution(value, "charts_resolution")
		assert.Equal(t, len(errors), 0)
	}
}

func TestValidateChartsResolutionNotAllowed(t *testing.T) {
	_, errors := validateChartsResolution("whatever", "charts_resolution")
	assert.Equal(t, len(errors), 1)
}
