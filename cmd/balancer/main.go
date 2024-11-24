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
	storage := repository.NewStorage(conf.Timeout, conf.Storages)
	vault := service.NewVault(file)
	upload := service.NewSplitUpload(file, storage)

	external, err := controller.NewBalancer(
		conf.Listen,
		conf.Limit,
		conf.Timeout,
		vault,
		upload,
	)
	graceful.Check(err)
	graceful.Add(external.Close)

	slog.Info("started", "listen", conf.Listen)
	graceful.Wait()
}
