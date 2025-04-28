package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var (
	// LogLevelNames maps log levels to their string representations
	LogLevelNames = map[LogLevel]string{
		LevelDebug:   "DEBUG",
		LevelInfo:    "INFO",
		LevelWarning: "WARNING",
		LevelError:   "ERROR",
		LevelFatal:   "FATAL",
	}

	// LogLevelEmojis maps log levels to their emoji representations
	LogLevelEmojis = map[LogLevel]string{
		LevelDebug:   "🔍",
		LevelInfo:    "ℹ️",
		LevelWarning: "⚠️",
		LevelError:   "❌",
		LevelFatal:   "💀",
	}
)

// Logger is a custom logger that includes file and line numbers
type Logger struct {
	*log.Logger
	level LogLevel
}

// New creates a new Logger instance
func New(level LogLevel) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", 0),
		level:  level,
	}
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo() string {
	// We need to skip:
	// 1. getCallerInfo
	// 2. formatMessage
	// 3. Debug/Info/Warning/Error/Fatal
	// 4. The actual caller
	// 5. The package-level function call
	skipFrames := 5

	pc := make([]uintptr, 15) // Increased buffer size
	n := runtime.Callers(0, pc)
	if n == 0 {
		return "unknown:0"
	}

	frames := runtime.CallersFrames(pc[:n])

	// Skip frames until we find the actual caller
	for i := 0; i < skipFrames; i++ {
		_, more := frames.Next()
		if !more {
			return "unknown:0"
		}
	}

	// Get the actual caller frame
	frame, _ := frames.Next()

	// Get just the filename without the full path
	_, filename := filepath.Split(frame.File)
	return fmt.Sprintf("%s:%d", filename, frame.Line)
}

// formatMessage formats the log message with emoji, level, file:line, and message
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	callerInfo := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s - %s",
		LogLevelEmojis[level],
		LogLevelNames[level],
		callerInfo,
		message,
	)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.Println(l.formatMessage(LevelDebug, format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.Println(l.formatMessage(LevelInfo, format, args...))
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.level <= LevelWarning {
		l.Println(l.formatMessage(LevelWarning, format, args...))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.Println(l.formatMessage(LevelError, format, args...))
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Println(l.formatMessage(LevelFatal, format, args...))
	os.Exit(1)
}

// Default logger instance
var defaultLogger = New(LevelInfo)

// SetLevel sets the log level for the default logger
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warning logs a warning message using the default logger
func Warning(format string, args ...interface{}) {
	defaultLogger.Warning(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}
