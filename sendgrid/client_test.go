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
// MockTransport mocks HTTP responses for testing purposes.
type MockTransport struct {
	Response *http.Response
	Error    error
}
// RoundTrip simulates the HTTP transport layer for testing.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Simulate network errors or responses
	return m.Response, m.Error
}
// TestContext holds the shared state and setup for the tests.
type TestContext struct {
	client           *sendgrid.Client
	originalTransport http.RoundTripper
	mockTransport    *MockTransport
}
// Initialize sets up the TestContext with the provided mock responses and status codes.
func (tc *TestContext) Initialize(apiKeys map[string]string, mockResponse *sendgrid.SendgridCreditsResponse, statusCode int, err error) {
	mockResponseBody, _ := json.Marshal(mockResponse)
	tc.mockTransport = &MockTransport{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(mockResponseBody)),
		},
		Error: err,
	}
	tc.client = sendgrid.NewClient(apiKeys)
	tc.originalTransport = http.DefaultTransport
	http.DefaultTransport = tc.mockTransport
}
// Cleanup restores the original HTTP transport after tests.
func (tc *TestContext) Cleanup() {
	http.DefaultTransport = tc.originalTransport
}
// TestFetchMetrics_Success tests successful fetching of metrics.
func TestFetchMetrics_Success(t *testing.T) {
	tc := &TestContext{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	mockResponse := sendgrid.SendgridCreditsResponse{
		Total:     1000,
		Remaining: 800,
		Used:      200,
		NextReset: "2024-12-01",
	}
	tc.Initialize(map[string]string{mockAccountName: mockAPIKey}, &mockResponse, http.StatusOK, nil)
	defer tc.Cleanup()
	metrics, statusCode, _, err := tc.client.FetchMetrics(mockAccountName)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.InDelta(t, 1000, metrics.Total, 0.001)
	assert.InDelta(t, 800, metrics.Remaining, 0.001)
	assert.InDelta(t, 200, metrics.Used, 0.001)
}
// TestFetchMetrics_AccountNotFound tests handling of an account not found error.
func TestFetchMetrics_AccountNotFound(t *testing.T) {
	tc := &TestContext{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	tc.Initialize(map[string]string{mockAccountName: mockAPIKey}, nil, http.StatusNotFound, nil)
	defer tc.Cleanup()
	metrics, statusCode, _, err := tc.client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, http.StatusNotFound, statusCode)
}
// TestFetchMetrics_HTTPError tests handling of HTTP errors.
func TestFetchMetrics_HTTPError(t *testing.T) {
	tc := &TestContext{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	tc.Initialize(map[string]string{mockAccountName: mockAPIKey}, nil, 0, http.ErrHandlerTimeout)
	defer tc.Cleanup()
	metrics, statusCode, _, err := tc.client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, 0, statusCode)
}
// TestFetchMetrics_Timeout tests handling of timeouts.
func TestFetchMetrics_Timeout(t *testing.T) {
	tc := &TestContext{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	tc.mockTransport = &MockTransport{
		Error: http.ErrHandlerTimeout, // Simulate a timeout error
	}
	tc.client = sendgrid.NewClient(map[string]string{
		mockAccountName: mockAPIKey,
	})
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = tc.mockTransport
	metrics, statusCode, _, err := tc.client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, 0, statusCode) 
}