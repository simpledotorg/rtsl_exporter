package dhis2

import (
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// sanitizeName prepares a valid prometheus metric name from a given URL
func sanitizeName(url string) string {
	// Remove the protocol part
	cleanURL := strings.TrimPrefix(url, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")

	// Replace dots and dashes with underscores
	cleanURL = strings.ReplaceAll(cleanURL, "-", "_")
	cleanURL = strings.ReplaceAll(cleanURL, ".", "_")

	return cleanURL
}

type Exporter struct {
	client *Client

	info *prometheus.GaugeVec
}

func NewExporter(client *Client) *Exporter {
	dynamicName := "system_info_" + sanitizeName(client.BaseURL)
	return &Exporter{
		client: client,
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "dhis2",
			Name:      dynamicName,
			Help:      "Information about the DHIS2 system",
		}, []string{"version", "revision", "contextPath", "buildTime"}),
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
		log.Printf("Failed to get dhis2 system information: %v\n", err)
		return // Early return on error to avoid using uninitialized info
	}

	// Set the version and revision as labels; gauge value is less meaningful here, just set to 1
	e.info.WithLabelValues(info.Version, info.Revision, info.ContextPath, info.BuildTime).Set(1)

	e.info.Collect(ch)
}
