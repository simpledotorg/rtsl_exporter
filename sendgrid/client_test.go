package sendgrid_test
import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/simpledotorg/rtsl_exporter/sendgrid"
)
// MockTransport mocks HTTP responses for testing
type MockTransport struct {
	Response *http.Response
	Error    error
}
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}
func TestFetchMetrics_Success(t *testing.T) {
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	mockResponse := sendgrid.SendgridCreditsResponse{
		Total:     1000,
		Remaining: 800,
		Used:      200,
		NextReset: "2024-12-01",
	}
	mockResponseBody, _ := json.Marshal(mockResponse)
	mockTransport := &MockTransport{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(mockResponseBody)),
		},
	}
	// Created a new Client and override the Transport field
	client := sendgrid.NewClient(map[string]string{
		mockAccountName: mockAPIKey,
	})
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = mockTransport
	metrics, statusCode, _, err := client.FetchMetrics(mockAccountName)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.InDelta(t, 1000, metrics.Total, 0.001)
	assert.InDelta(t, 800, metrics.Remaining, 0.001)
	assert.InDelta(t, 200, metrics.Used, 0.001)
}
func TestFetchMetrics_AccountNotFound(t *testing.T) {
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	mockTransport := &MockTransport{
		Response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		},
	}
	client := sendgrid.NewClient(map[string]string{
		mockAccountName: mockAPIKey,
	})
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = mockTransport
	metrics, statusCode, _, err := client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, http.StatusNotFound, statusCode)
}
func TestFetchMetrics_HTTPError(t *testing.T) {
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	mockTransport := &MockTransport{
		Error: http.ErrHandlerTimeout,
	}
	// Create a new Client and override the Transport field
	client := sendgrid.NewClient(map[string]string{
		mockAccountName: mockAPIKey,
	})
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = mockTransport
	metrics, statusCode, _, err := client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, 0, statusCode)
}
