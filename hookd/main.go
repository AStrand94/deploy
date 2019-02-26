package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/navikt/deployment/common/pkg/logging"
	"github.com/navikt/deployment/hookd/pkg/config"
	"github.com/navikt/deployment/hookd/pkg/github"
	"github.com/navikt/deployment/hookd/pkg/secrets"
	"github.com/navikt/deployment/hookd/pkg/server"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
)

var cfg = config.DefaultConfig()

func init() {
	flag.StringVar(&cfg.ListenAddress, "listen-address", cfg.ListenAddress, "IP:PORT")
	flag.StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "Log format, either 'json' or 'text'.")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging verbosity level.")
	flag.StringVar(&cfg.WebhookURL, "webhook-url", cfg.WebhookURL, "Externally available URL to events endpoint.")
	flag.IntVar(&cfg.ApplicationID, "app-id", cfg.ApplicationID, "Github App ID.")
	flag.IntVar(&cfg.InstallID, "install-id", cfg.InstallID, "Github App installation ID.")
	flag.StringVar(&cfg.KeyFile, "key-file", cfg.KeyFile, "Path to PEM key owned by Github App.")
	flag.StringVar(&cfg.VaultAddress, "vault-address", cfg.VaultAddress, "Address to Vault HTTP API.")
	flag.StringVar(&cfg.VaultPath, "vault-path", cfg.VaultPath, "Base path to hookd data in Vault.")
	flag.StringSliceVar(&cfg.KafkaBrokers, "kafka-brokers", cfg.KafkaBrokers, "Comma-separated list of Kafka brokers, HOST:PORT.")
	flag.StringVar(&cfg.KafkaTopic, "kafka-topic", cfg.KafkaTopic, "Kafka topic for deployd communication.")
}

func run() error {
	flag.Parse()

	if err := logging.Setup(cfg.LogLevel, cfg.LogFormat); err != nil {
		return err
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if len(vaultToken) == 0 {
		return fmt.Errorf("the VAULT_TOKEN environment variable needs to be set")
	}

	secretClient, err := secrets.New(cfg.VaultAddress, vaultToken, cfg.VaultPath)
	if err != nil {
		return fmt.Errorf("while configuring secret client: %s", err)
	}

	log.Info("hookd is starting")

	kafka, err := sarama.NewSyncProducer(cfg.KafkaBrokers, nil)
	if err != nil {
		return fmt.Errorf("while configuring Kafka: %s", err)
	}

	githubClient, err := github.ApplicationClient(cfg.ApplicationID, cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("cannot instantiate Github installation client: %s", err)
	}

	installationClient, err := github.InstallationClient(cfg.ApplicationID, cfg.InstallID, cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("cannot instantiate Github installation client: %s", err)
	}

	baseHandler := server.Handler{
		Config:                   *cfg,
		SecretClient:             secretClient,
		KafkaProducer:            kafka,
		KafkaTopic:               cfg.KafkaTopic,
		GithubClient:             githubClient,
		GithubInstallationClient: installationClient,
	}
	http.Handle("/register/repository", &server.LifecycleHandler{Handler: baseHandler})
	http.Handle("/events", &server.DeploymentHandler{Handler: baseHandler})
	srv := &http.Server{
		Addr: cfg.ListenAddress,
	}
	return srv.ListenAndServe()
}

func main() {
	err := run()
	if err != nil {
		log.Errorf("Fatal error: %s", err)
		os.Exit(1)
	}
}
