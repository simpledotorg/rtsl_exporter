package dhis2

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	clients []*Client
	info    *prometheus.GaugeVec
	mutex   sync.Mutex
}

func NewExporter(clients []*Client) *Exporter {
	return &Exporter{
		clients: clients,
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "dhis2",
			Name:      "system_info",
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
	var wg sync.WaitGroup

	for _, client := range e.clients {
		wg.Add(1)

		go func(client *Client) {
			defer wg.Done()

			info, err := client.GetInfo()
			if err != nil {
				log.Printf("Failed to get system information from %s: %v\n", client.BaseURL, err)
				return
			}

			e.mutex.Lock()
			e.info.WithLabelValues(info.Version, info.Revision, info.ContextPath, info.BuildTime).Set(1)
			e.mutex.Unlock()
		}(client)
	}

	wg.Wait()
	e.info.Collect(ch)
}
