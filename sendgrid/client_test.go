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
// MockTransport mocks HTTP responses for testing.
type MockTransport struct {
	Response *http.Response
	Error    error
}
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}
// ClientTester encapsulates the state and methods for testing the sendgrid client.
type ClientTester struct {
	client           *sendgrid.Client
	originalTransport http.RoundTripper
	mockTransport    *MockTransport
}
// Initialize sets up the ClientTester with mock responses and status codes.
func (ct *ClientTester) Initialize(apiKeys map[string]string, mockResponse *sendgrid.SendgridCreditsResponse, statusCode int, err error) {
	mockResponseBody, _ := json.Marshal(mockResponse)
	ct.mockTransport = &MockTransport{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(mockResponseBody)),
		},
		Error: err,
	}
	ct.client = sendgrid.NewClient(apiKeys)
	ct.originalTransport = http.DefaultTransport
	http.DefaultTransport = ct.mockTransport
}
// Cleanup restores the original HTTP transport after tests.
func (ct *ClientTester) Cleanup() {
	http.DefaultTransport = ct.originalTransport
}
// assertSuccess verifies successful metrics collection.
func (ct *ClientTester) assertSuccess(t *testing.T, expectedResponse sendgrid.SendgridCreditsResponse, mockAccountName string) {
	metrics, statusCode, _, err := ct.client.FetchMetrics(mockAccountName)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.InDelta(t, expectedResponse.Total, metrics.Total, 0.001)
	assert.InDelta(t, expectedResponse.Remaining, metrics.Remaining, 0.001)
	assert.InDelta(t, expectedResponse.Used, metrics.Used, 0.001)
}
// assertAccountNotFound verifies the account not found scenario.
func (ct *ClientTester) assertAccountNotFound(t *testing.T, mockAccountName string) {
	metrics, statusCode, _, err := ct.client.FetchMetrics(mockAccountName)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, http.StatusNotFound, statusCode)
}
// assertHTTPError verifies the HTTP error scenario.
func (ct *ClientTester) assertHTTPError(t *testing.T) {
	metrics, statusCode, _, err := ct.client.FetchMetrics("mockAccount")
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Equal(t, 0, statusCode)
}
func TestFetchMetrics_Success(t *testing.T) {
	ct := &ClientTester{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	mockResponse := sendgrid.SendgridCreditsResponse{
		Total:     1000,
		Remaining: 800,
		Used:      200,
		NextReset: "2024-12-01",
	}
	ct.Initialize(map[string]string{mockAccountName: mockAPIKey}, &mockResponse, http.StatusOK, nil)
	defer ct.Cleanup()
	ct.assertSuccess(t, mockResponse, mockAccountName)
}
func TestFetchMetrics_AccountNotFound(t *testing.T) {
	ct := &ClientTester{}
	mockAPIKey := "mockAPIKey"
	mockAccountName := "mockAccount"
	ct.Initialize(map[string]string{mockAccountName: mockAPIKey}, nil, http.StatusNotFound, nil)
	defer ct.Cleanup()
	ct.assertAccountNotFound(t, mockAccountName)
}
func TestFetchMetrics_HTTPError(t *testing.T) {
	ct := &ClientTester{}
	mockAPIKey := "mockAPIKey"
	ct.Initialize(map[string]string{mockAPIKey: mockAPIKey}, nil, 0, http.ErrHandlerTimeout)
	defer ct.Cleanup()
	ct.assertHTTPError(t)
}