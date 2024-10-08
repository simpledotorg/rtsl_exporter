package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/simpledotorg/rtsl_exporter/alphasms"
	"github.com/simpledotorg/rtsl_exporter/dhis2"
	"github.com/simpledotorg/rtsl_exporter/sendgrid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	ALPHASMSAPIKey   string                   `yaml:"alphasms_api_key"`
	SendGridAccounts []sendgrid.AccountConfig `yaml:"sendgrid_accounts"`
	DHIS2Endpoints   []struct {
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
func gracefulShutdown(server *http.Server) {
	// Create a context with a timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Notify when the shutdown process is complete
	idleConnectionsClosed := make(chan struct{})
	go func() {
		defer close(idleConnectionsClosed)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
	}()
	// Wait for the server to shut down
	select {
	case <-ctx.Done():
		log.Println("HTTP Server Shutdown Timeout")
	case <-idleConnectionsClosed:
		log.Println("HTTP Server Shutdown Complete")
	}
}
func main() {
	log.SetFlags(0)
	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	// Alphasms
	if config.ALPHASMSAPIKey == "" {
		log.Fatalf("ALPHASMS_API_KEY not provided in config file")
	}
	alphasmsClient := alphasms.Client{APIKey: config.ALPHASMSAPIKey}
	alphasmsExporter := alphasms.NewExporter(&alphasmsClient)
	prometheus.MustRegister(alphasmsExporter)
	// DHIS2
	dhis2Clients := []*dhis2.Client{}
	for _, endpoint := range config.DHIS2Endpoints {
		dhis2Client := dhis2.Client{
			Username:          endpoint.Username,
			Password:          endpoint.Password,
			BaseURL:           endpoint.BaseURL,
			ConnectionTimeout: dhis2.DefaultConnectionTimeout,
		}
		dhis2Clients = append(dhis2Clients, &dhis2Client)
	}
	dhis2Exporter := dhis2.NewExporter(dhis2Clients)
	prometheus.MustRegister(dhis2Exporter)
	// Register SendGrid exporters with time zones
	sendGridConfigMap := make(map[string]sendgrid.AccountConfig)
	for _, account := range config.SendGridAccounts {
		sendGridConfigMap[account.AccountName] = sendgrid.AccountConfig{
			AccountName: account.AccountName,
			APIKey:      account.APIKey,
			TimeZone:    account.TimeZone,
		}
	}
	sendgridExporter := sendgrid.NewExporter(sendGridConfigMap)
	prometheus.MustRegister(sendgridExporter)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting server on :8080")
	httpServer := &http.Server{
		Addr: ":8080",
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("Shutdown signal received")
		gracefulShutdown(httpServer)
	}()
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}
	log.Println("Bye bye")
}
