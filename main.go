package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/alphasms_exporter/alphasms"
	"github.com/simpledotorg/alphasms_exporter/dhis2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ALPHASMSAPIKey string `yaml:"alphasms_api_key"`
	DHIS2Endpoints []struct {
		BaseURL  string `yaml:"base_url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"dhis2_endpoints"`
}

func readConfig(configPath string) (*Config, error) {
	config := &Config{}
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if config.ALPHASMSAPIKey == "" {
		log.Fatalf("ALPHASMS_API_KEY not provided in config file")
	}
	alphasmsClient := alphasms.Client{APIKey: config.ALPHASMSAPIKey}
	alphasmsExporter := alphasms.NewExporter(&alphasmsClient)
	prometheus.MustRegister(alphasmsExporter)

	for _, endpoint := range config.DHIS2Endpoints {
		dhis2Client := dhis2.Client{
			Username: endpoint.Username,
			Password: endpoint.Password,
			BaseURL:  endpoint.BaseURL,
		}
		dhis2Exporter := dhis2.NewExporter(&dhis2Client)
		prometheus.MustRegister(dhis2Exporter)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
