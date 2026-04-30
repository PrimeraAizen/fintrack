package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

// Level represents the logging level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Config holds logger configuration
type Config struct {
	Level       Level  `mapstructure:"level"`
	Format      string `mapstructure:"format"` // json, text
	Output      string `mapstructure:"output"` // stdout, stderr, file
	FilePath    string `mapstructure:"file_path"`
	AddSource   bool   `mapstructure:"add_source"`
	Service     string `mapstructure:"service"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	config *Config
}

// Fields represents key-value pairs for structured logging
type Fields map[string]interface{}

// New creates a new logger instance
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = &Config{
			Level:       LevelInfo,
			Format:      "json",
			Output:      "stdout",
			AddSource:   false,
			Service:     "app",
			Version:     "1.0.0",
			Environment: "development",
		}
	}

	var output io.Writer
	switch config.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		if config.FilePath == "" {
			return nil, fmt.Errorf("file path is required when output is 'file'")
		}
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	default:
		output = os.Stdout
	}

	// Convert level string to slog.Level
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	// Create logger with service context
	logger := slog.New(handler).With(
		slog.String("service", config.Service),
		slog.String("version", config.Version),
		slog.String("environment", config.Environment),
	)

	return &Logger{
		Logger: logger,
		config: config,
	}, nil
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields Fields) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
	}
}

// WithContext creates a new logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context fields
	fields := Fields{}

	// Add request ID if present
	if reqID := ctx.Value("request_id"); reqID != nil {
		fields["request_id"] = reqID
	}

	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	// Add correlation ID if present
	if corrID := ctx.Value("correlation_id"); corrID != nil {
		fields["correlation_id"] = corrID
	}

	if len(fields) > 0 {
		return l.WithFields(fields)
	}

	return l
}

// WithError creates a new logger with error field
func (l *Logger) WithError(err error) *Logger {
	return l.WithFields(Fields{"error": err.Error()})
}

// WithComponent creates a new logger with component field
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithFields(Fields{"component": component})
}

// WithOperation creates a new logger with operation field
func (l *Logger) WithOperation(operation string) *Logger {
	return l.WithFields(Fields{"operation": operation})
}

// WithDuration creates a new logger with duration field
func (l *Logger) WithDuration(duration time.Duration) *Logger {
	return l.WithFields(Fields{"duration_ms": duration.Milliseconds()})
}

// WithRequest creates a new logger with HTTP request fields
func (l *Logger) WithRequest(method, path string) *Logger {
	return l.WithFields(Fields{
		"http_method": method,
		"http_path":   path,
	})
}

// WithResponse creates a new logger with HTTP response fields
func (l *Logger) WithResponse(statusCode int, size int64) *Logger {
	return l.WithFields(Fields{
		"http_status":   statusCode,
		"response_size": size,
	})
}

// WithDatabase creates a new logger with database operation fields
func (l *Logger) WithDatabase(operation, table string) *Logger {
	return l.WithFields(Fields{
		"db_operation": operation,
		"db_table":     table,
	})
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
	panic(fmt.Sprintf(msg, args...))
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(ctx context.Context, method, path, userAgent string, duration time.Duration, statusCode int) {
	l.WithContext(ctx).
		WithRequest(method, path).
		WithDuration(duration).
		WithResponse(statusCode, 0).
		WithFields(Fields{"user_agent": userAgent}).
		Info("HTTP request completed")
}

// LogDatabase logs a database operation
func (l *Logger) LogDatabase(ctx context.Context, operation, table string, duration time.Duration, err error) {
	logger := l.WithContext(ctx).WithDatabase(operation, table).WithDuration(duration)

	if err != nil {
		logger.WithError(err).Error("Database operation failed")
	} else {
		logger.Info("Database operation completed")
	}
}

// LogBusiness logs a business operation
func (l *Logger) LogBusiness(ctx context.Context, operation string, duration time.Duration, err error) {
	logger := l.WithContext(ctx).WithOperation(operation).WithDuration(duration)

	if err != nil {
		logger.WithError(err).Error("Business operation failed")
	} else {
		logger.Info("Business operation completed")
	}
}

// GetCallerInfo returns information about the calling function
func (l *Logger) GetCallerInfo() Fields {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return Fields{}
	}

	funcName := runtime.FuncForPC(pc).Name()
	// Extract just the function name without package path
	parts := strings.Split(funcName, ".")
	funcName = parts[len(parts)-1]

	return Fields{
		"caller_file": file,
		"caller_line": line,
		"caller_func": funcName,
	}
}

// SetGlobal sets this logger as the global logger
func (l *Logger) SetGlobal() {
	slog.SetDefault(l.Logger)
}

// Close closes the logger (useful for file-based logging)
func (l *Logger) Close() error {
	// For file-based logging, we might need to close the file
	// This is a placeholder for future implementation
	return nil
}

// Default creates a default logger instance
func Default() *Logger {
	config := &Config{
		Level:       LevelInfo,
		Format:      "json",
		Output:      "stdout",
		AddSource:   false,
		Service:     "app",
		Version:     "1.0.0",
		Environment: "development",
	}

	logger, _ := New(config)
	return logger
}

// NewFromEnv creates a logger from environment variables
func NewFromEnv() (*Logger, error) {
	config := &Config{
		Level:       Level(os.Getenv("LOG_LEVEL")),
		Format:      getEnvOrDefault("LOG_FORMAT", "json"),
		Output:      getEnvOrDefault("LOG_OUTPUT", "stdout"),
		FilePath:    os.Getenv("LOG_FILE_PATH"),
		AddSource:   getEnvOrDefault("LOG_ADD_SOURCE", "false") == "true",
		Service:     getEnvOrDefault("SERVICE_NAME", "app"),
		Version:     getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}

	return New(config)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
