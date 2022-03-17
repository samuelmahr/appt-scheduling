package configuration

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Log                            log.Logger
	DatabaseURL                    string
	PostgresMaxOpenConns           int
	PostgresMaxIdleConns           int
	PostgresMaxConnLifetimeSeconds int

	TestDatabaseURL string
}

var globalConfig AppConfig

func Configuration() *AppConfig {
	return &globalConfig
}

func Configure() (*AppConfig, error) {
	viper.SetConfigName("local")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	maxConns, err := strconv.Atoi(getEnv("postgres_max_open_conns", "10"))
	if err != nil {
		log.Fatal(err)
	}

	maxIdle, err := strconv.Atoi(getEnv("postgres_max_idle_conns", "5"))
	if err != nil {
		log.Fatal(err)
	}

	maxLifetime, err := strconv.Atoi(getEnv("postgres_max_conn_lifetime_seconds", "3600"))
	if err != nil {
		log.Fatal(err)
	}

	c := AppConfig{}
	c.DatabaseURL = getEnv("database_url", "")
	c.PostgresMaxOpenConns = maxConns
	c.PostgresMaxIdleConns = maxIdle
	c.PostgresMaxConnLifetimeSeconds = maxLifetime
	c.TestDatabaseURL = getEnv("test_database_url", "")

	c.Log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	c.Log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	c.Log.SetLevel(log.WarnLevel)

	return &c, nil
}

func getEnv(key, fallback string) string {
	value := viper.GetString(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}
