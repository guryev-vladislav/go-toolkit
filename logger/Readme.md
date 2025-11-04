# Logger Library

A unified logging library for Go applications that supports multiple logging backends (Zap and Slog) with a consistent API.

## Features

- **Dual Backend Support**: Choose between Zap (high-performance) or Slog (standard library)
- **Structured Logging**: Support for structured logs with key-value pairs
- **Multiple Outputs**: Console and file output with different formats
- **Context Support**: Built-in context propagation for distributed tracing
- **SQL Error Helpers**: Specialized methods for database operation errors
- **Flexible Configuration**: Easy configuration via struct
- **Production Ready**: Includes panic recovery, graceful shutdown, and proper error handling

## Installation

Add the library to your Go module:

```bash
go get github.com/guryev-vladislav/digital-showcase/golang/lib/logger