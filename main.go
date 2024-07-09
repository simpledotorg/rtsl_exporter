package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/alphasms_exporter/alphasms"
)

func main() {
	apikey := os.Getenv("ALPHASMS_API_KEY")
	if apikey == "" {
		log.Fatalf("Failed to load ALPHASMS_API_KEY from environment variable")
	}

	client := alphasms.Client{APIKey: apikey}

	exporter := alphasms.NewExporter(&client)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
