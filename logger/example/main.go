package main

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/zap"

	slg "github.com/guryev-vladislav/digital-showcase/golang/lib/logger/pkg/slog_logger"
	zlg "github.com/guryev-vladislav/digital-showcase/golang/lib/logger/pkg/zap_logger"
)

// runWithZapLogger demonstrates usage of Zap logger
func runWithZapLogger(logger *zlg.ZapLogger) {
	defer logger.End()

	logger.Info("Application started")

	err := errors.New("database connection failed")
	logger.Error("Operation failed", zap.Error(err))

	logger.Info("Application finished")
}

// runWithSlogLogger demonstrates usage of Slog logger
func runWithSlogLogger(logger *slg.Logger) {
	defer logger.End()

	logger.Info("Application started")

	err := errors.New("database connection failed")
	logger.Error("Operation failed", slog.String("error", err.Error()))

	logger.Info("Application finished")
}

func main() {
	// Using Zap logger
	configZap := zlg.Config{
		ServiceName: "example-service", // Service name
		Version:     "1.0.0",           // Service version
		LogLevel:    "info",            // Logging level
	}
	zapFactory, err := zlg.NewZapLoggerFactory(configZap)
	if err != nil {
		panic(err)
	}
	defer zapFactory.Close()

	ctx := context.Background()
	zapLogger := zapFactory.GetLogger(ctx, zap.String("logger_type", "zap"), zap.String("request_id", "abc123"))
	runWithZapLogger(zapLogger)

	// Using Slog logger
	configSlog := slg.Config{
		ServiceName: "example-service", // Service name
		Version:     "1.0.0",           // Service version
		LogLevel:    "info",            // Logging level
	}
	slogFactory, err := slg.NewLoggerFactory(configSlog)
	if err != nil {
		panic(err)
	}
	defer slogFactory.Close()

	slogLogger := slogFactory.GetLogger(ctx, slog.String("logger_type", "slog"), slog.String("request_id", "xyz789"))
	runWithSlogLogger(slogLogger)
}
