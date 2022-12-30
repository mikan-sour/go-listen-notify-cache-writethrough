package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const UT_ENV = "../../.env"

func TestNewConfigFail(t *testing.T) {
	_, err := New("./.unit-test-fail.env")

	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if !strings.Contains(err.Error(), "error loading .env file:") {
		t.Errorf("wanted a string containing `error loading .env file:` but got %s", err.Error())
	}

}

func TestDnsSuccess(t *testing.T) {
	cfg, err := New(UT_ENV)
	if err != nil {
		t.Fatal()
	}
	expected := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	dns := cfg.Dns()

	if dns != expected {
		t.Errorf("expected %s but got %s", expected, dns)
	}

}

func TestGetCacheConnDetailsSuccess(t *testing.T) {
	cfg, err := New(UT_ENV)
	if err != nil {
		t.Fatal()
	}

	expected := &CacheConnDetails{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("CACHE_HOST"), os.Getenv("CACHE_PORT")),
		Password: os.Getenv("CACHE_PASSWORD"),
	}

	res := cfg.GetCacheConnDetails()

	if res.Addr != expected.Addr {
		t.Fatalf("addresses don't match - expected %s, got %s", expected.Addr, res.Addr)
	}

	if res.Password != expected.Password {
		t.Fatalf("passwords don't match - expected %s, got %s", expected.Password, res.Password)
	}
}
