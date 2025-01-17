package sendgrid

import (
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"testing"
)

func TestExporterCollect(t *testing.T) {
	// Activate the HTTP mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	
	// Mock a successful SendGrid API response for the main account
	httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/user/credits",
		httpmock.NewStringResponder(200, `{
			"total": 1000,
			"remain": 800,
			"used": 200,
			"next_reset": "2024-02-20"
		}`))

	// Mock a successful SendGrid API response for a subuser
	httpmock.RegisterResponder("GET", "https://api.sendgrid.com/v3/subusers/mockSubuser/credits",
		httpmock.NewStringResponder(200, `{
			"total": 500,
			"remain": 300,
			"used": 200
		}`))

	// Created a new Exporter with subuser configuration
	accountConfigs := map[string]AccountConfig{
		"mockAccount": {
			AccountName: "mockAccount",
			APIKey:      "mockAPIKey",
			TimeZone:    "UTC",
			Subusers: []SubuserConfig{
				{
					SubuserName: "mockSubuser",
					APIKey:      "mockSubuserAPIKey",
					TimeZone:    "UTC",
				},
			},
		},
	}

	exporter := NewExporter(accountConfigs)
	t.Run("Successful metrics collection for main account and subuser", func(t *testing.T) {
		expectedMetrics := []string{
			"sendgrid_email_limit_count",          // main account and subuser
			"sendgrid_email_remaining_count",      // main account and subuser
			"sendgrid_email_used_count",           // main account and subuser
			"sendgrid_monitoring_http_return_code", // main account and subuser
			"sendgrid_monitoring_http_response_time_seconds", // main account and subuser
		}
		
		// Collect and count the metrics
		count := testutil.CollectAndCount(exporter, expectedMetrics...)
		expectedCount := len(expectedMetrics) * 2
		if count != expectedCount {
			t.Errorf("expected %d metrics, but got %d", expectedCount, count)
		}
	})
}
