package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/alphasms_exporter/alphasms"
	"github.com/simpledotorg/alphasms_exporter/dhis2"
)

func main() {
	apikey := os.Getenv("ALPHASMS_API_KEY")
	if apikey == "" {
		log.Fatalf("Failed to load ALPHASMS_API_KEY from environment variable")
	}
	client := alphasms.Client{APIKey: apikey}
	exporter := alphasms.NewExporter(&client)

	dhis2Client := dhis2.Client{Username: "", Password: ""}
	dhis2exporter := dhis2.NewExporter(&dhis2Client)

	prometheus.MustRegister(exporter)
	prometheus.MustRegister(dhis2exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
