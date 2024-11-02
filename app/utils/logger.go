package utils

import (
	"log/slog"
	"os"
)

func NewLogger(component string) *slog.Logger {
	options := &slog.HandlerOptions{}
	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stdout, options)
	return slog.New(handler)
}
