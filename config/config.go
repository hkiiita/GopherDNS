package config

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type DomainConfig struct {
	DomainIpMapping map[string]string `yaml:"domains"`
}

var globalDomainConfiguration DomainConfig

func SetDomainConfig() error {
	return readConfig()

}

func GetDomainConfig() *DomainConfig {
	return &globalDomainConfiguration
}

func readConfig() error {
	// Load the YAML configuration
	file, err := os.Open("domains.yaml")
	if err != nil {
		log.Fatalf("Error opening YAML file: %v", err)
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&globalDomainConfiguration)
	if err != nil {
		log.Fatalf("Error decoding YAML file: %v", err)
		return err
	}
	return nil
}
