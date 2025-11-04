package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	lg "github.com/guryev-vladislav/digital-showcase/golang/lib/logger/pkg"
)

const (
	unknownFunctionName = "unknown function name"
	funcNameSeparator   = ": "
)

type SQLErrorType int

const (
	SQLSelect SQLErrorType = iota
	SQLInsert
	SQLUpdate
	SQLDelete
)

type Logger struct {
	slogLog      *slog.Logger
	functionName string
	ctx          context.Context
}

type LoggerFactory struct {
	slogLog *slog.Logger
	config  Config
	file    *os.File
}

type Config struct {
	ServiceName string
	Version     string
	LogLevel    string
	OutputPath  string
}

func NewLoggerFactory(cfg Config) (*LoggerFactory, error) {
	slogLog, file, err := newSlogLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &LoggerFactory{
		slogLog: slogLog,
		config:  cfg,
		file:    file,
	}, nil
}

func (f *LoggerFactory) GetLogger(ctx context.Context, attrs ...slog.Attr) *Logger {
	msg := lg.MsgStart
	if len(attrs) > 0 {
		msg = lg.MsgStartWithParams
	}

	functionName := getFunctionName(2, true)

	logger := &Logger{
		slogLog:      f.slogLog,
		functionName: functionName,
		ctx:          ctx,
	}

	if len(attrs) > 0 {
		logger.slogLog.InfoContext(ctx, createMessageWithFuncName(functionName, msg), attrsToArgs(attrs)...)
	} else {
		logger.slogLog.InfoContext(ctx, createMessageWithFuncName(functionName, msg))
	}

	return logger
}

func (f *LoggerFactory) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

func (l *Logger) WithFields(attrs ...slog.Attr) *Logger {
	return &Logger{
		slogLog:      l.slogLog.With(attrsToArgs(attrs)...),
		functionName: l.functionName,
		ctx:          l.ctx,
	}
}

func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	if l.slogLog.Enabled(l.ctx, slog.LevelDebug) {
		l.slogLog.DebugContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(attrs)...)
	}
}

func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	if l.slogLog.Enabled(l.ctx, slog.LevelInfo) {
		l.slogLog.InfoContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(attrs)...)
	}
}

func (l *Logger) Warning(msg string, attrs ...slog.Attr) {
	if l.slogLog.Enabled(l.ctx, slog.LevelWarn) {
		l.slogLog.WarnContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(attrs)...)
	}
}

func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	if l.slogLog.Enabled(l.ctx, slog.LevelError) {
		l.slogLog.ErrorContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(attrs)...)
	}
}

func (l *Logger) ErrorIn(funcName string, err error, attrs ...slog.Attr) {
	msg := fmt.Sprintf(lg.MsgCompletesWithError, funcName)
	allAttrs := append(attrs, slog.String("error", err.Error()))
	l.slogLog.ErrorContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(allAttrs)...)
}

func (l *Logger) ErrorSQL(operation SQLErrorType, table string, err error, attrs ...slog.Attr) {
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

	allAttrs := append(attrs, slog.String("error", err.Error()), slog.String("table", table))
	l.slogLog.ErrorContext(l.ctx, createMessageWithFuncName(l.functionName, msg), attrsToArgs(allAttrs)...)
}

func (l *Logger) ErrorSQLSelect(table string, err error, attrs ...slog.Attr) {
	l.ErrorSQL(SQLSelect, table, err, attrs...)
}

func (l *Logger) ErrorSQLInsert(table string, err error, attrs ...slog.Attr) {
	l.ErrorSQL(SQLInsert, table, err, attrs...)
}

func (l *Logger) ErrorSQLUpdate(table string, err error, attrs ...slog.Attr) {
	l.ErrorSQL(SQLUpdate, table, err, attrs...)
}

func (l *Logger) ErrorSQLDelete(table string, err error, attrs ...slog.Attr) {
	l.ErrorSQL(SQLDelete, table, err, attrs...)
}

func (l *Logger) End() {
	if err := recover(); err != nil {
		l.slogLog.ErrorContext(l.ctx, lg.MsgPanicWasCatched,
			slog.Any("error", err),
			slog.String("function", l.functionName))
		l.slogLog.InfoContext(l.ctx, createMessageWithFuncName(l.functionName, lg.MsgEnd))
		panic(err)
	}
	l.slogLog.InfoContext(l.ctx, createMessageWithFuncName(l.functionName, lg.MsgEnd))
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

func attrsToArgs(attrs []slog.Attr) []any {
	if len(attrs) == 0 {
		return nil
	}
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	return args
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func sourceKeyReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if source, ok := a.Value.Any().(*slog.Source); ok {
			filename := filepath.Base(source.File)
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filename, source.Line))
		}
	}
	return a
}

func newSlogLogger(cfg Config) (*slog.Logger, *os.File, error) {
	logLevel := parseLogLevel(cfg.LogLevel)

	var file *os.File
	var handlers []slog.Handler

	consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       logLevel,
		AddSource:   true,
		ReplaceAttr: sourceKeyReplaceAttr,
	})
	handlers = append(handlers, NewCallerHandler(consoleHandler, 4))

	if cfg.OutputPath != "" {
		var err error
		file, err = os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level:       logLevel,
			AddSource:   true,
			ReplaceAttr: sourceKeyReplaceAttr,
		})
		handlers = append(handlers, NewCallerHandler(fileHandler, 4))
	}

	handler := NewMultiHandler(handlers...)
	if handler == nil {
		return slog.Default(), file, nil
	}

	return slog.New(handler), file, nil
}
