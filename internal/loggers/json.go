package loggers

import (
	"log/slog"
	"os"
)

var AppLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
