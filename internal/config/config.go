package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

const defaultRunAddress = "localhost:8080"
const defaultDatabaseURI = "postgresql://gophermart_app:qazxswedc@localhost:5432/gophermart-loyalty"

type Configuration struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	BaseURL              string
}

func NewConfiguration() *Configuration {
	cfg := new(Configuration)

	cfg.fillFromFlags()

	err := cfg.fillFromEnvironment()
	if err != nil {
		log.Println(err)
	}
	/*
		runAddress := []rune(cfg.RunAddress)
		if runAddress[len(runAddress)-1] != '/' {
			cfg.RunAddress += "/"
		}
	*/
	cfg.BaseURL = "http://" + cfg.RunAddress + "/"

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
