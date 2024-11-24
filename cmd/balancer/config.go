package main

import (
	"balancer/pkg/validation"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Listen   string        `env:"LISTEN"   validate:"required"`
	Limit    int           `env:"LIMIT"    validate:"min=4000,max=20000000000"`
	Timeout  time.Duration `env:"TIMEOUT"  validate:"min=0s,max=120m"`
	Dir      string        `env:"DIR"      validate:"required"`
	Storages []string      `env:"STORAGES" validate:"required"`
}

func NewConfig() (c Config, e error) {
	if err := godotenv.Load("balancer.env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load env variables from file: %w", err)
	}

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return Config{}, fmt.Errorf("parse env variables to config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&c); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", validation.Pretty(err))
	}

	return c, nil
}
