package dhis2

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client *Client

	info *prometheus.GaugeVec
}

func NewExporter(client *Client) *Exporter {
	return &Exporter{
		client: client,
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "dhis2",
			Name:      "system_info",
			Help:      "Information about the DHIS2 system",
		}, []string{"version", "revision"}),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.info.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	info, err := e.client.GetInfo()
	if err != nil {
		log.Printf("Failed to get system information: %v\n", err)
		return // Early return on error to avoid using uninitialized info
	}

	// Set the version and revision as labels; gauge value is less meaningful here, just set to 1
	e.info.WithLabelValues(info.Version, info.Revision).Set(1)

	e.info.Collect(ch)
}
