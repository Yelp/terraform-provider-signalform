package signalform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatePlotTypeTimeChartAllowed(t *testing.T) {
	for _, value := range []string{"LineChart", "AreaChart", "ColumnChart", "Histogram"} {
		_, errors := validatePlotTypeTimeChart(value, "plot_type")
		assert.Equal(t, len(errors), 0)
	}
}

func TestValidatePlotTypeTimeChartNotAllowed(t *testing.T) {
	_, errors := validatePlotTypeTimeChart("absolute", "plot_type")
	assert.Equal(t, len(errors), 1)
}
