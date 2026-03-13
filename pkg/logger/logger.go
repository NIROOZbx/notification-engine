package logger

import (
	"io"
	"os"
	"time"

	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(cfg *config.LogConfig) zerolog.Logger {

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   true,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}

	level, _ := zerolog.ParseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	return zerolog.New(io.MultiWriter(consoleWriter, fileWriter)).
		With().
		Timestamp().
		Caller().
		Logger()
}