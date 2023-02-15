package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

const defaultRunAddress = "http://localhost:8080/"
const defaultDatabaseURI = "postgresql://gopher_app:qazxswedc@localhost:5432/gopher-loyalty"

type Configuration struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfiguration() *Configuration {
	cfg := new(Configuration)

	cfg.fillFromFlags()

	err := cfg.fillFromEnvironment()
	if err != nil {
		log.Println(err)
	}

	runAddress := []rune(cfg.RunAddress)
	if runAddress[len(runAddress)-1] != '/' {
		cfg.RunAddress += "/"
	}

	databaseURI := []rune(cfg.DatabaseURI)
	if databaseURI[len(databaseURI)-1] != '/' {
		cfg.DatabaseURI += "/"
	}

	log.Println("Resulting config:", cfg)

	return cfg
}

func (c *Configuration) fillFromEnvironment() error {
	err := env.Parse(c)
	if err != nil {
		return err
	}

	log.Println("Environment config:", c)

	return nil
}

func (c *Configuration) fillFromFlags() {
	flag.StringVar(&c.RunAddress, "a", defaultRunAddress, "string with server address")
	flag.StringVar(&c.DatabaseURI, "d", defaultDatabaseURI, "string with database URI")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "string with database URI")

	flag.Parse()

	log.Println("Console flags:", c)
}
