package models

import "os"

const (
	envAccessTokenKey = "ACCESS_TOKEN"
	envDomainKey      = "DOMAIN"
	outputDir         = "output"
)

type Config struct {
	AccessToken string
	Domain      string
}

func NewConfig() *Config {
	accessToken := os.Getenv(envAccessTokenKey)
	domain := os.Getenv(envDomainKey)

	return &Config{
		AccessToken: accessToken,
		Domain:      domain,
	}
}
