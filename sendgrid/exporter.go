package sendgrid

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client *Client
	emailLimit      *prometheus.GaugeVec
	emailsRemaining *prometheus.GaugeVec
	emailUsed       *prometheus.GaugeVec
	planExpiration  *prometheus.GaugeVec
	httpReturnCode  *prometheus.GaugeVec
	httpResponseTime *prometheus.GaugeVec
}

func NewExporter(apiKeys map[string]string) *Exporter {
	return &Exporter{
		client: NewClient(apiKeys),
		emailLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_limit_count",
			Help:      "The total email limit for the account.",
		}, []string{"account_name"}),
		emailsRemaining: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "emails_remaining_count",
			Help:      "The number of emails remaining for the account.",
		}, []string{"account_name"}),
		emailUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_used_count",
			Help:      "The number of emails used for the account.",
		}, []string{"account_name"}),
		planExpiration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "plan_expiration_seconds",
			Help:      "The time until the plan expires, in seconds.",
		}, []string{"account_name"}),
		httpReturnCode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "monitoring_http_return_code",
			Help:      "The HTTP return code from the SendGrid API request.",
		}, []string{"account_name"}),
		httpResponseTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "monitoring_http_response_time_seconds",
			Help:      "The response time of the SendGrid API request, in seconds.",
		}, []string{"account_name"}),
	}
}

// Describe sends the descriptions of the metrics to Prometheus.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.emailLimit.Describe(ch)
	e.emailsRemaining.Describe(ch)
	e.emailUsed.Describe(ch)
	e.planExpiration.Describe(ch)
	e.httpReturnCode.Describe(ch)
	e.httpResponseTime.Describe(ch)
}

// Collect retrieves metrics and sends them to Prometheus.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// Collect metrics for each account
	for accountName := range e.client.APIKeys {
		metrics, statusCode, responseTime, err := e.client.FetchMetrics(accountName)
		if err != nil {
			log.Printf("Failed to get metrics for account %s: %v", accountName, err)
			continue
		}

		// Set metrics values for each account
		e.emailLimit.WithLabelValues(accountName).Set(metrics.Total)
		e.emailsRemaining.WithLabelValues(accountName).Set(metrics.Remaining)
		e.emailUsed.WithLabelValues(accountName).Set(metrics.Used)

		// Parse the plan expiration date
		dateFormats := []string{
			time.RFC3339,
			"2006-01-02",
		}

		var planResetDate time.Time
		var parseErr error
		for _, format := range dateFormats {
			planResetDate, parseErr = time.Parse(format, metrics.NextReset)
			if parseErr == nil {
				break
			}
		}

		if parseErr != nil {
			log.Printf("Failed to parse plan reset date: %v", parseErr)
			continue
		}

		planResetDate = planResetDate.Add(24 * time.Hour)
		timeUntilExpiration := planResetDate.Sub(time.Now()).Seconds()
		if timeUntilExpiration < 0 {
			timeUntilExpiration = 0
		}

		e.planExpiration.WithLabelValues(accountName).Set(timeUntilExpiration)
		e.httpReturnCode.WithLabelValues(accountName).Set(float64(statusCode))
		e.httpResponseTime.WithLabelValues(accountName).Set(responseTime.Seconds())
	}

	// Collect all metrics once
	e.emailLimit.Collect(ch)
	e.emailsRemaining.Collect(ch)
	e.emailUsed.Collect(ch)
	e.planExpiration.Collect(ch)
	e.httpReturnCode.Collect(ch)
	e.httpResponseTime.Collect(ch)
}
