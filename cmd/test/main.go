package main

import (
	"balancer/internal/repository"
	"balancer/internal/service"
	"balancer/pkg/logger"
	"flag"
	"log/slog"
	"time"
)

func main() {
	addr := flag.String("a", "0.0.0.0:8080", "balancer address without proto")
	mode := flag.String("m", "upload", "upload or download mode")
	name := flag.String("n", "file.txt", "name or path for both modes")
	dir := flag.String("d", "data", "dir for calculating meta")
	timeout := flag.Duration("t", 120*time.Second, "request and response timeout like 300s or 2h45m")
	flag.Parse()

	slog.SetDefault(logger.New())
	file := repository.NewFile(*dir)
	balancer := repository.NewBalancer(*addr, *timeout)
	upload := service.NewPlainUpload(file, balancer)

	switch *mode {
	case "upload":
		if err := upload.Upload(*name); err != nil {
			slog.Error("upload", "error", err)
		}
	case "download":
	default:
		flag.PrintDefaults()
	}
}
