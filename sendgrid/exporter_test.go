package sendgrid
import (
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"testing"
)
// TestExporterCollect tests the Exporter struct using httpmock
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
	// Create a new Exporter
	accountNames := map[string]string{
		"mockAccount": "mockAPIKey",
	}
	exporter := NewExporter(accountNames)
	t.Run("Successful metrics collection", func(t *testing.T) {
		// Define expected metric names
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
}