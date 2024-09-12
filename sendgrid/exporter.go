package sendgrid

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client         *Client
	timeZones      map[string]*time.Location // Added timeZones map
	emailLimit     *prometheus.GaugeVec
	emailRemaining *prometheus.GaugeVec
	emailUsed      *prometheus.GaugeVec
	planExpiration *prometheus.GaugeVec
	httpReturnCode *prometheus.GaugeVec
	httpResponseTime *prometheus.GaugeVec
}

func NewExporter(apiKeys map[string]string, timeZones map[string]*time.Location) *Exporter {
	return &Exporter{
		client: NewClient(apiKeys),
		timeZones: timeZones, // Added timeZones to Exporter
		emailLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_limit_count",
			Help:      "The total email limit for the account.",
		}, []string{"account_name"}),
		emailRemaining: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_remaining_count",
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
	e.emailRemaining.Describe(ch)
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
		e.emailRemaining.WithLabelValues(accountName).Set(metrics.Remaining)
		e.emailUsed.WithLabelValues(accountName).Set(metrics.Used)
		// Load the time zone for the account
		timeZone, exists := e.timeZones[accountName]
		if !exists {
			timeZone = time.UTC // Default to UTC if the time zone is not provided
		}
		// Parse the plan expiration date
		dateFormat := "2006-01-02"
		planResetDate, parseErr := time.ParseInLocation(dateFormat, metrics.NextReset, timeZone)
		if parseErr != nil {
			log.Printf("Failed to parse plan reset date for account %s: %v", accountName, parseErr)
			continue
		}
		currentTime := time.Now().In(timeZone)	
		// Calculate time until expiration
		timeUntilExpiration := planResetDate.Sub(currentTime).Seconds()

		log.Printf("timeUntilExpiration: %+v", timeUntilExpiration)
		e.planExpiration.WithLabelValues(accountName).Set(timeUntilExpiration)
		e.httpReturnCode.WithLabelValues(accountName).Set(float64(statusCode))
		e.httpResponseTime.WithLabelValues(accountName).Set(responseTime.Seconds())
	}
	// Collect all metrics once
	e.emailLimit.Collect(ch)
	e.emailRemaining.Collect(ch)
	e.emailUsed.Collect(ch)
	e.planExpiration.Collect(ch)
	e.httpReturnCode.Collect(ch)
	e.httpResponseTime.Collect(ch)
}
