package signalform

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `Test Response`)
	}))
	defer server.Close()

	status_code, body, err := sendRequest("GET", server.URL, "token", nil)
	assert.Equal(t, 200, status_code)
	assert.Equal(t, "Test Response\n", string(body))
	assert.Nil(t, err)
}

func TestSendRequestResponseNotFound(t *testing.T) {
	// Handler returns 404 page not found
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	status_code, body, err := sendRequest("POST", server.URL, "token", nil)
	assert.Equal(t, 404, status_code)
	assert.Contains(t, string(body), "page not found")
	assert.Nil(t, err)
}

func TestSendRequestFail(t *testing.T) {
	// Client will fail to send due to invalid URL
	status_code, body, err := sendRequest("GET", "", "token", nil)
	assert.Equal(t, -1, status_code)
	assert.Nil(t, body)
	assert.Contains(t, err.Error(), "Failed sending GET request")
}

func TestSendRequestTimeout(t *testing.T) {
	timeout := time.Duration(1 * time.Second)
	server := httptest.NewServer(http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout)
	}), timeout, "Timeout occurred"))
	defer server.Close()

	status_code, body, err := sendRequest("POST", server.URL, "token", nil)
	assert.Equal(t, 503, status_code)
	assert.Equal(t, "Timeout occurred", string(body))
	assert.Nil(t, err)
}

func TestValidateTimeSpanTypeAbsolute(t *testing.T) {
	_, errors := validateTimeSpanType("absolute", "time_span_type")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeSpanTypeRelative(t *testing.T) {
	_, errors := validateTimeSpanType("relative", "time_span_type")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeSpanTypeNotAllowed(t *testing.T) {
	_, errors := validateTimeSpanType("foo", "time_span_type")
	assert.Equal(t, 1, len(errors))
}

func TestValidateTimeEpoch(t *testing.T) {
	_, errors := validateTime("621129601", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeSecondsNotAllowed(t *testing.T) {
	_, errors := validateTime("621129", "time_start")
	assert.Equal(t, 1, len(errors))
}

func TestValidateTimeStringSeconds(t *testing.T) {
	_, errors := validateTime("-5s", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeStringMinutes(t *testing.T) {
	_, errors := validateTime("-15m", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeStringHours(t *testing.T) {
	_, errors := validateTime("-5h", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeStringDays(t *testing.T) {
	_, errors := validateTime("-50d", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeStringWeeks(t *testing.T) {
	_, errors := validateTime("-5w", "time_start")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeStringEnd(t *testing.T) {
	_, errors := validateTime("-5d", "time_end")
	assert.Equal(t, 1, len(errors))
}

func TestValidateTimeNowEnd(t *testing.T) {
	_, errors := validateTime("Now", "time_end")
	assert.Equal(t, 0, len(errors))
}

func TestValidateTimeNowStart(t *testing.T) {
	_, errors := validateTime("Now", "time_start")
	assert.Equal(t, 1, len(errors))
}

func TestValidateTimeStringMonthsNotAllowed(t *testing.T) {
	_, errors := validateTime("-5M", "time_start")
	assert.Equal(t, 1, len(errors))
}

func TestValidateTimeStringNotAllowed(t *testing.T) {
	_, errors := validateTime("foo", "time_start")
	assert.Equal(t, 1, len(errors))
}
