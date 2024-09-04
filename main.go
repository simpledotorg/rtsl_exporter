package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/rtsl_exporter/alphasms"
	"github.com/simpledotorg/rtsl_exporter/dhis2"
	"github.com/simpledotorg/rtsl_exporter/sendgrid"
)

type Config struct {
	ALPHASMSAPIKey string `yaml:"alphasms_api_key"`
	SendGridAccounts []struct {
		AccountName string `yaml:"account_name"`
		APIKey      string `yaml:"api_key"`
	} `yaml:"sendgrid_accounts"`
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

	// Register SendGrid exporters
	apiKeys := make(map[string]string)
	for _, account := range config.SendGridAccounts {
		apiKeys[account.AccountName] = account.APIKey
	}
	sendgridExporter := sendgrid.NewExporter(apiKeys)
	prometheus.MustRegister(sendgridExporter)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	http.ListenAndServe(":8080", nil)
}
