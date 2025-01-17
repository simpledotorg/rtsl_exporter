package sendgrid

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

type Exporter struct {
	client           *Client
	timeZones        map[string]*time.Location // Map of time locations
	emailLimit       *prometheus.GaugeVec
	emailRemaining   *prometheus.GaugeVec
	emailUsed        *prometheus.GaugeVec
	planExpiration   *prometheus.GaugeVec
	httpReturnCode   *prometheus.GaugeVec
	httpResponseTime *prometheus.GaugeVec
	accounts         map[string]AccountConfig
}

type AccountConfig struct {
	AccountName string         `yaml:"account_name"`
	APIKey      string         `yaml:"api_key"`
	TimeZone    string         `yaml:"time_zone"`
	Subusers    []SubuserConfig `yaml:"subusers"`
}

type SubuserConfig struct {
	SubuserName string `yaml:"subuser_name"`
	APIKey      string `yaml:"api_key"`
	TimeZone    string `yaml:"time_zone"`
}

func NewExporter(accounts map[string]AccountConfig) *Exporter {
	apiKeys := make(map[string]string)
	timeZones := make(map[string]*time.Location)
	for accountName, accountConfig := range accounts {
		apiKeys[accountName] = accountConfig.APIKey
		loc, err := time.LoadLocation(accountConfig.TimeZone)
		if err != nil {
			log.Printf("Error loading time zone for account %s: %v", accountName, err)
			loc = time.UTC // Default to UTC if time zone cannot be loaded
		}
		timeZones[accountName] = loc
	}

	return &Exporter{
		client:    NewClient(apiKeys),
		timeZones: timeZones,
		accounts:  accounts,
		emailLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_limit_count",
			Help:      "The total email limit for the account.",
		}, []string{"account_name", "account_type", "parent_account_name"}),
		emailRemaining: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_remaining_count",
			Help:      "The number of emails remaining for the account.",
		}, []string{"account_name", "account_type", "parent_account_name"}),
		emailUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "email_used_count",
			Help:      "The number of emails used for the account.",
		}, []string{"account_name", "account_type", "parent_account_name"}),
		planExpiration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "plan_expiration_seconds",
			Help:      "The time until the plan expires, in seconds.",
		}, []string{"account_name", "account_type"}),
		httpReturnCode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "monitoring_http_return_code",
			Help:      "The HTTP return code from the SendGrid API request.",
		}, []string{"account_name", "account_type", "parent_account_name"}),
		httpResponseTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sendgrid",
			Name:      "monitoring_http_response_time_seconds",
			Help:      "The response time of the SendGrid API request, in seconds.",
		}, []string{"account_name", "account_type", "parent_account_name"}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.emailLimit.Describe(ch)
	e.emailRemaining.Describe(ch)
	e.emailUsed.Describe(ch)
	e.planExpiration.Describe(ch)
	e.httpReturnCode.Describe(ch)
	e.httpResponseTime.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for accountName, accountConfig := range e.accounts {
		log.Printf("Processing metrics for account: %s", accountName)

		// Fetch metrics for the main account
		metrics, statusCode, responseTime, err := e.client.FetchMetrics(accountName)
		if err != nil {
			log.Printf("Failed to get metrics for account %s: %v", accountName, err)
			continue
		}
		e.emailLimit.WithLabelValues(accountName, "user", accountName).Set(metrics.Total)
		e.emailRemaining.WithLabelValues(accountName, "user", accountName).Set(metrics.Remaining)
		e.emailUsed.WithLabelValues(accountName, "user", accountName).Set(metrics.Used)
		timeZone := e.timeZones[accountName]
		dateFormat := "2006-01-02"
		planResetDate, parseErr := time.ParseInLocation(dateFormat, metrics.NextReset, timeZone)
		if parseErr != nil {
			log.Printf("Failed to parse plan reset date for account %s: %v", accountName, parseErr)
			continue
		}
		currentTime := time.Now().In(timeZone)
		timeUntilExpiration := planResetDate.Sub(currentTime).Seconds()
		e.planExpiration.WithLabelValues(accountName, "user").Set(timeUntilExpiration)
		e.httpReturnCode.WithLabelValues(accountName, "user", accountName).Set(float64(statusCode))
		e.httpResponseTime.WithLabelValues(accountName, "user", accountName).Set(responseTime.Seconds())

		// Process metrics for subusers
		subusers := accountConfig.Subusers
		if subusers == nil || len(subusers) == 0 {
			continue
		}

		for _, subuser := range subusers {
			subuserMetrics, subuserStatusCode, subuserResponseTime, err := e.client.FetchMetricsForSubuser(subuser.SubuserName, accountName)
			if err != nil {
				log.Printf("Failed to get metrics for subuser %s: %v", subuser.SubuserName, err)
				continue
			}
			e.emailLimit.WithLabelValues(subuser.SubuserName, "subuser", accountName).Set(subuserMetrics.Total)
			e.emailRemaining.WithLabelValues(subuser.SubuserName, "subuser", accountName).Set(subuserMetrics.Remaining)
			e.emailUsed.WithLabelValues(subuser.SubuserName, "subuser", accountName).Set(subuserMetrics.Used)

			subuserPlanResetDate, subuserParseErr := time.ParseInLocation(dateFormat, subuserMetrics.NextReset, timeZone)
			if subuserParseErr == nil {
				timeUntilExpirationSubuser := subuserPlanResetDate.Sub(currentTime).Seconds()
				e.planExpiration.WithLabelValues(subuser.SubuserName, "subuser").Set(timeUntilExpirationSubuser)
			}

			e.httpReturnCode.WithLabelValues(subuser.SubuserName, "subuser", accountName).Set(float64(subuserStatusCode))
			e.httpResponseTime.WithLabelValues(subuser.SubuserName, "subuser", accountName).Set(subuserResponseTime.Seconds())
		}
	}

	e.emailLimit.Collect(ch)
	e.emailRemaining.Collect(ch)
	e.emailUsed.Collect(ch)
	e.planExpiration.Collect(ch)
	e.httpReturnCode.Collect(ch)
	e.httpResponseTime.Collect(ch)
}
