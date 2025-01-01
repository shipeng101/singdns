package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Global logger instance
	Logger *zap.SugaredLogger
)

// LogConfig represents logger configuration
type LogConfig struct {
	Level      string `json:"level"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"` // number of backups
	MaxAge     int    `json:"max_age"`     // days
	Compress   bool   `json:"compress"`
	Console    bool   `json:"console"` // output to console
}

// InitLogger initializes the global logger
func InitLogger(config *LogConfig) error {
	// Create log directory if not exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("create log directory failed: %v", err)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("parse log level failed: %v", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create writers
	var writers []zapcore.WriteSyncer

	// File writer with rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	})
	writers = append(writers, fileWriter)

	// Console writer
	if config.Console {
		consoleWriter := zapcore.AddSync(os.Stdout)
		writers = append(writers, consoleWriter)
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writers...),
		level,
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Logger = logger.Sugar()

	return nil
}

// LogError logs an error with context
func LogError(err error, context string, fields ...interface{}) {
	if Logger != nil {
		Logger.Errorw(context, append([]interface{}{"error", err}, fields...)...)
	}
}

// LogInfo logs an info message with context
func LogInfo(message string, fields ...interface{}) {
	if Logger != nil {
		Logger.Infow(message, fields...)
	}
}

// LogWarning logs a warning message with context
func LogWarning(message string, fields ...interface{}) {
	if Logger != nil {
		Logger.Warnw(message, fields...)
	}
}

// LogDebug logs a debug message with context
func LogDebug(message string, fields ...interface{}) {
	if Logger != nil {
		Logger.Debugw(message, fields...)
	}
}

// RequestLogger returns a middleware that logs HTTP requests
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			rw := &responseWriter{w, http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r)

			// Log request
			duration := time.Since(start)
			LogInfo("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", duration,
				"ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, message string, details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// WriteJSON writes JSON response with proper error handling
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		LogError(err, "Failed to encode JSON response")
	}
}
