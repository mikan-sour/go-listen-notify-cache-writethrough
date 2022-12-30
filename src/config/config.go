package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const ENV = ".env"

type Config interface {
	Dns() string
	GetCacheConnDetails() CacheConnDetails
}

type ConfigImpl struct {
	DB_Host        string
	DB_Port        string
	DB_Name        string
	DB_Username    string
	DB_Password    string
	Cache_Host     string
	Cache_Port     string
	Cache_Password string
}

type CacheConnDetails struct {
	Addr     string
	Password string
}

func New(envFile string) (Config, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %s", err.Error())
	}

	return &ConfigImpl{
		DB_Host:        os.Getenv("DB_HOST"),
		DB_Port:        os.Getenv("DB_PORT"),
		DB_Name:        os.Getenv("DB_NAME"),
		DB_Password:    os.Getenv("DB_PASSWORD"),
		DB_Username:    os.Getenv("DB_USERNAME"),
		Cache_Host:     os.Getenv("CACHE_HOST"),
		Cache_Port:     os.Getenv("CACHE_PORT"),
		Cache_Password: os.Getenv("CACHE_PASSWORD"),
	}, nil
}

func (cfg *ConfigImpl) Dns() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DB_Host, cfg.DB_Port, cfg.DB_Username, cfg.DB_Password, cfg.DB_Name)
}

func (cfg *ConfigImpl) GetCacheConnDetails() CacheConnDetails {
	return CacheConnDetails{
		Addr:     fmt.Sprintf("%s:%s", cfg.Cache_Host, cfg.Cache_Port),
		Password: cfg.Cache_Password,
	}
}
