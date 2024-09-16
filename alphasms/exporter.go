package alphasms

import (
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client *Client

	balance prometheus.Gauge
	error   prometheus.Gauge
	date    prometheus.Gauge
}

func NewExporter(client *Client) *Exporter {
	return &Exporter{
		client: client,
		balance: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "alphasms",
			Name:      "user_balance_amount",
			Help:      "The current balance amount.",
		}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "alphasms",
			Name:      "user_balance_error",
			Help:      "The current error code while connecting to api.",
		}),
		date: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "alphasms",
			Name:      "user_balance_validity",
			Help:      "Validity date of balance amount.",
		}),
	}
}

// Descriptors
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.balance.Describe(ch)
	e.error.Describe(ch)
	e.date.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	apiResp, balanceData, err := e.client.GetUserBalance()
	if err != nil {
		log.Printf("Failed to get GetUserBalance: %v\n", err)
		return
	}

	balance, err := strconv.ParseFloat(balanceData.Balance, 64)
	if err != nil {
		log.Printf("Failed to parse balance string: %v\n", err)
	}

	e.balance.Set(balance)
	e.error.Set(float64(apiResp.Error))

	// Convert date string to unix timestamp
	t, err := time.Parse("2006-01-02 00:00:00", balanceData.Validity)
	if err != nil {
		log.Printf("Failed to parse date string: %v\n", err)
	}

	e.date.Set(float64(t.Unix()))

	e.balance.Collect(ch)
	e.error.Collect(ch)
	e.date.Collect(ch)
}
