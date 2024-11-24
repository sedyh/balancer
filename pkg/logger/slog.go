package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

func New() *slog.Logger {
	return slog.New(
		console.NewHandler(
			os.Stderr,
			&console.HandlerOptions{
				Level:      slog.LevelDebug,
				TimeFormat: time.TimeOnly,
			},
		),
	)
}
