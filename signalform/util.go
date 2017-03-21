package signalform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// Workaround for Signalfx bug related to post processing and lastUpdatedTime
	OFFSET = 10000.0
)

/*
  Utility function that wraps http calls to SignalFx
*/
func sendRequest(method string, url string, token string, payload []byte) (int, []byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-SF-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return -1, nil, fmt.Errorf("Failed sending %s request to Signalfx: %s", method, err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("Failed reading response body from %s request: %s", method, err.Error())
	}

	return resp.StatusCode, body, nil
}

/*
  Validates the time_span_type field against a list of allowed words.
*/
func validateTimeSpanType(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "relative" && value != "absolute" {
		errors = append(errors, fmt.Errorf("%s not allowed; must be either relative or absolute", value))
	}
	return
}

/*
  Validates the plot_type field against a list of allowed words.
*/
func validatePlotTypeTimeChart(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if value != "LineChart" && value != "AreaChart" && value != "ColumnChart" && value != "Histogram" {
		errors = append(errors, fmt.Errorf("%s not allowed; Must be \"LineChart\", \"AreaChart\", \"ColumnChart\", or \"Histogram\"", value))
	}
	return
}
