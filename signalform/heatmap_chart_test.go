package signalform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateHeatmapChartColors(t *testing.T) {
	_, err := validateHeatmapChartColor("blue", "color")
	assert.Equal(t, 0, len(err))
}

func TestValidateHeatmapChartColorsFail(t *testing.T) {
	_, err := validateHeatmapChartColor("whatever", "color")
	assert.Equal(t, 1, len(err))
}
