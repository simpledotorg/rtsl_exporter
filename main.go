package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"io/ioutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/rtsl_exporter/alphasms"
	"github.com/simpledotorg/rtsl_exporter/dhis2"
	"github.com/simpledotorg/rtsl_exporter/sendgrid"
	"gopkg.in/yaml.v2"
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
	log.SetFlags(0)

	var listenAddress = flag.String("listen", ":8080", "Listen address.")
	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

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
	log.Printf("Starting server on %s", *listenAddress)

	httpServer := http.Server{
		Addr: *listenAddress,
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(idleConnectionsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-idleConnectionsClosed

	log.Printf("Bye bye")
}
