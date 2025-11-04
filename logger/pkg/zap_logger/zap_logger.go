package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	lg "github.com/guryev-vladislav/digital-showcase/golang/lib/logger/pkg"
)

const (
	unknownFunctionName  = "unknown function name"
	serviceNameSeparator = " - "
	funcNameSeparator    = ": "
)

type SQLErrorType int

const (
	SQLSelect SQLErrorType = iota
	SQLInsert
	SQLUpdate
	SQLDelete
)

type ZapLogger struct {
	zapLog       *zap.Logger
	functionName string
}

type ZapLoggerFactory struct {
	zapLog *zap.Logger
	config Config
}

type Config struct {
	ServiceName string
	Version     string
	LogLevel    string
	OutputPath  string
}

var (
	consoleEncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	jsonEncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

func NewZapLoggerFactory(cfg Config) (*ZapLoggerFactory, error) {
	zapLog, err := newZapLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &ZapLoggerFactory{
		zapLog: zapLog,
		config: cfg,
	}, nil
}

func (f *ZapLoggerFactory) GetLogger(_ context.Context, fields ...zap.Field) *ZapLogger {
	msg := lg.MsgStart
	if len(fields) > 0 {
		msg = lg.MsgStartWithParams
	}

	functionName := getFunctionName(1, true)

	logger := &ZapLogger{
		zapLog:       f.zapLog,
		functionName: functionName,
	}

	logger.zapLog.Info(createMessageWithFuncName(functionName, msg), fields...)

	return logger
}

func (f *ZapLoggerFactory) Close() error {
	return f.zapLog.Sync()
}

func createMessageWithFuncName(funcName, msg string) string {
	if funcName == "" {
		return msg
	}
	return funcName + funcNameSeparator + msg
}

func getFunctionName(skippedStackFrames int, shortFunctionName bool) string {
	pc, _, _, ok := runtime.Caller(skippedStackFrames + 1)
	if !ok {
		return unknownFunctionName
	}

	funcName := runtime.FuncForPC(pc).Name()

	if shortFunctionName {
		if idx := strings.LastIndex(funcName, "/"); idx != -1 {
			funcName = funcName[idx+1:]
		}
		if idx := strings.LastIndex(funcName, "."); idx != -1 {
			funcName = funcName[idx+1:]
		}
	}

	return funcName
}

func customEncodeCaller(serviceName, version string) zapcore.CallerEncoder {
	var prefix string
	switch {
	case serviceName == "" && version != "":
		prefix = version
	case serviceName != "" && version != "":
		prefix = serviceName + " " + version
	case serviceName != "":
		prefix = serviceName
	default:
		return zapcore.ShortCallerEncoder
	}

	return func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(prefix + serviceNameSeparator + caller.TrimmedPath())
	}
}

func parseZapLogLevel(level string) (zapcore.Level, error) {
	if level == "" {
		return zapcore.DebugLevel, nil
	}
	minLogLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return zapcore.DebugLevel, err
	}
	return minLogLevel, nil
}

func newZapLogger(cfg Config) (*zap.Logger, error) {
	logLevel, err := parseZapLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	consoleCfg := consoleEncoderConfig
	consoleCfg.EncodeCaller = customEncodeCaller(cfg.ServiceName, cfg.Version)

	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)
	jsonEncoder := zapcore.NewJSONEncoder(jsonEncoderConfig)

	cores := []zapcore.Core{
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stderr),
			zap.NewAtomicLevelAt(logLevel),
		),
	}

	if cfg.OutputPath != "" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		fileCore := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(file),
			zap.NewAtomicLevelAt(logLevel),
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

func (z *ZapLogger) WithFields(fields ...zap.Field) *ZapLogger {
	return &ZapLogger{
		zapLog:       z.zapLog.With(fields...),
		functionName: z.functionName,
	}
}

func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	if z.zapLog.Core().Enabled(zapcore.DebugLevel) {
		z.zapLog.Debug(createMessageWithFuncName(z.functionName, msg), fields...)
	}
}

func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	if z.zapLog.Core().Enabled(zapcore.InfoLevel) {
		z.zapLog.Info(createMessageWithFuncName(z.functionName, msg), fields...)
	}
}

func (z *ZapLogger) Warning(msg string, fields ...zap.Field) {
	if z.zapLog.Core().Enabled(zapcore.WarnLevel) {
		z.zapLog.Warn(createMessageWithFuncName(z.functionName, msg), fields...)
	}
}

func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	if z.zapLog.Core().Enabled(zapcore.ErrorLevel) {
		z.zapLog.Error(createMessageWithFuncName(z.functionName, msg), fields...)
	}
}

func (z *ZapLogger) ErrorIn(funcName string, err error, fields ...zap.Field) {
	msg := fmt.Sprintf(lg.MsgCompletesWithError, funcName)
	allFields := append(fields, zap.Error(err))
	z.zapLog.Error(createMessageWithFuncName(z.functionName, msg), allFields...)
}

func (z *ZapLogger) ErrorSQL(operation SQLErrorType, table string, err error, fields ...zap.Field) {
	var msg string
	switch operation {
	case SQLSelect:
		msg = fmt.Sprintf(lg.MsgSQLSelectWithError, table)
	case SQLInsert:
		msg = fmt.Sprintf(lg.MsgSQLInsertWithError, table)
	case SQLUpdate:
		msg = fmt.Sprintf(lg.MsgSQLUpdateWithError, table)
	case SQLDelete:
		msg = fmt.Sprintf(lg.MsgSQLDeleteWithError, table)
	default:
		msg = fmt.Sprintf("SQL operation error on table %s", table)
	}

	allFields := append(fields, zap.Error(err))
	z.zapLog.Error(createMessageWithFuncName(z.functionName, msg), allFields...)
}

func (z *ZapLogger) ErrorSQLSelect(table string, err error, fields ...zap.Field) {
	z.ErrorSQL(SQLSelect, table, err, fields...)
}

func (z *ZapLogger) ErrorSQLInsert(table string, err error, fields ...zap.Field) {
	z.ErrorSQL(SQLInsert, table, err, fields...)
}

func (z *ZapLogger) ErrorSQLUpdate(table string, err error, fields ...zap.Field) {
	z.ErrorSQL(SQLUpdate, table, err, fields...)
}

func (z *ZapLogger) ErrorSQLDelete(table string, err error, fields ...zap.Field) {
	z.ErrorSQL(SQLDelete, table, err, fields...)
}

func (z *ZapLogger) Panic(msg string, fields ...zap.Field) {
	z.zapLog.Panic(createMessageWithFuncName(z.functionName, msg), fields...)
}

func (z *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	z.zapLog.Fatal(createMessageWithFuncName(z.functionName, msg), fields...)
}

func (z *ZapLogger) End() {
	if err := recover(); err != nil {
		z.zapLog.Error(lg.MsgPanicWasCatched,
			zap.Any("error", err),
			zap.Stack("stacktrace"))

		z.zapLog.Info(createMessageWithFuncName(z.functionName, lg.MsgEnd))
		panic(err)
	}

	z.zapLog.Info(createMessageWithFuncName(z.functionName, lg.MsgEnd))
}
