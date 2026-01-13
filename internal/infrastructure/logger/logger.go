// Package logger provides file-based logging for the application.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger provides file-based logging.
type Logger struct {
	file   *os.File
	logger *log.Logger
	path   string
	mu     sync.Mutex
}

var (
	global     *Logger
	globalOnce sync.Once
)

// DefaultLogPath returns the default log file path.
func DefaultLogPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.ExpandEnv("$HOME/.config")
	}
	// Use date-based log file name
	date := time.Now().Format("2006-01-02")
	return filepath.Join(configDir, "yukti", "logs", fmt.Sprintf("yukti-%s.log", date))
}

// Init initializes the global logger.
func Init() error {
	var initErr error
	globalOnce.Do(func() {
		global, initErr = New(DefaultLogPath())
	})
	return initErr
}

// New creates a new logger writing to the specified file.
func New(path string) (*Logger, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open file for appending
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	return &Logger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags),
		path:   path,
	}, nil
}

// Path returns the log file path.
func (l *Logger) Path() string {
	return l.path
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[INFO] "+format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[ERROR] "+format, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[DEBUG] "+format, args...)
}

// Close closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Writer returns an io.Writer for the log file.
func (l *Logger) Writer() io.Writer {
	return l.file
}

// Global returns the global logger instance.
// Must call Init() first.
func Global() *Logger {
	return global
}

// Info logs to the global logger.
func Info(format string, args ...any) {
	if global != nil {
		global.Info(format, args...)
	}
}

// Error logs to the global logger.
func Error(format string, args ...any) {
	if global != nil {
		global.Error(format, args...)
	}
}

// Debug logs to the global logger.
func Debug(format string, args ...any) {
	if global != nil {
		global.Debug(format, args...)
	}
}

// Path returns the global log file path.
func Path() string {
	if global != nil {
		return global.Path()
	}
	return DefaultLogPath()
}

// Close closes the global logger.
func Close() error {
	if global != nil {
		return global.Close()
	}
	return nil
}
