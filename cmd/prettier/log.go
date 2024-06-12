package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func colorize(colorCode int, v string, noColor bool) string {
	if noColor {
		return v
	}
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

type handler struct {
	level   slog.Level
	noColor bool
}

var _ slog.Handler = (*handler)(nil)

// Enabled implements slog.Handler.
func (h handler) Enabled(_ context.Context, l slog.Level) bool {
	return l.Level() >= h.level.Level()
}

// Handle implements slog.Handler.
func (h handler) Handle(_ context.Context, r slog.Record) error {
	var color int
	switch r.Level {
	case slog.LevelDebug:
		color = blue
	case slog.LevelInfo:
		// Default
	case slog.LevelWarn:
		color = yellow
	case slog.LevelError:
		color = red
	}

	level := strings.ToLower(r.Level.String())
	if color != 0 {
		level = colorize(color, level, h.noColor)
	}

	fmt.Fprintf(os.Stderr, "[%s] %s\n", level, r.Message)

	return nil
}

// WithAttrs implements slog.Handler.
func (h handler) WithAttrs([]slog.Attr) slog.Handler {
	// We don't use attrs in this program.
	return h
}

// WithGroup implements slog.Handler.
func (h handler) WithGroup(string) slog.Handler {
	// We don't use groups in this program.
	return h
}
