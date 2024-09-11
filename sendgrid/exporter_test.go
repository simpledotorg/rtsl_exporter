package sendgrid
import (
	"net/http"
	"testing"
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)
func TestExporterCollect(t *testing.T) {
	// Activate the HTTP mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Mock a successful SendGrid API response
	httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
		httpmock.NewStringResponder(200, `{
			"total": 1000,
			"remain": 800,
			"used": 200,
			"next_reset": "2024-02-20"
		}`))
	// Created a new Exporter
	accountNames := map[string]string{
		"mockAccount": "mockAPIKey",
	}
	exporter := NewExporter(accountNames)
	t.Run("Successful metrics collection", func(t *testing.T) {
		expectedMetrics := []string{
			"sendgrid_email_limit_count",
			"sendgrid_email_remaining_count",
			"sendgrid_email_used_count",
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
			"sendgrid_plan_expiration_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
	t.Run("HTTP error response", func(t *testing.T) {
		// Mock an error response from the SendGrid API
		httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
			httpmock.NewStringResponder(500, `{
				"error": "Internal Server Error"
			}`))
		expectedMetrics := []string{
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
	t.Run("Timeout or no response", func(t *testing.T) {
		// Mock a timeout from the SendGrid API
		httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
			httpmock.NewErrorResponder(http.ErrHandlerTimeout))
		expectedMetrics := []string{
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
	t.Run("Invalid JSON response", func(t *testing.T) {
		// Mock an invalid JSON response
		httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
			httpmock.NewStringResponder(200, `{
				"total": 1000,
				"remain": "INVALID",
				"used": 200
			}`))
		expectedMetrics := []string{
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
	t.Run("Partial data in API response", func(t *testing.T) {
		// Mock a response with missing fields (e.g., no 'used' field)
		httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
			httpmock.NewStringResponder(200, `{
				"total": 1000,
				"remain": 800
			}`))
		expectedMetrics := []string{
			"sendgrid_email_limit_count",
			"sendgrid_email_remaining_count",
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
	t.Run("Unauthorized API response", func(t *testing.T) {
		// Mock a 401 Unauthorized response from the SendGrid API
		httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
			httpmock.NewStringResponder(401, `{
				"error": "Unauthorized"
			}`))
		expectedMetrics := []string{
			"sendgrid_monitoring_http_return_code",
			"sendgrid_monitoring_http_response_time_seconds",
		}
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics)
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
}