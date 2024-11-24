package main

import (
	"balancer/internal/controller"
	"balancer/internal/repository"
	"balancer/internal/service"
	"balancer/pkg/graceful"
	"balancer/pkg/logger"
	"log/slog"
)

func main() {
	slog.SetDefault(logger.New())

	conf, err := NewConfig()
	graceful.Check(err)

	file := repository.NewFile(conf.Dir)
	vault := service.NewVault(file)

	external, err := controller.NewStorage(
		conf.Listen,
		conf.Limit,
		conf.Timeout,
		vault,
	)
	graceful.Check(err)
	graceful.Add(external.Close)

	slog.Info("started", "listen", conf.Listen)
	graceful.Wait()
}
