package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Level представляет уровень логирования
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger предоставляет функциональность для логирования
type Logger struct {
	level  Level
	logger *log.Logger
}

// New создает новый экземпляр Logger с указанным уровнем
func New(levelStr string) *Logger {
	level := parseLevel(levelStr)

	logger := log.New(os.Stdout, "", 0)

	return &Logger{
		level:  level,
		logger: logger,
	}
}

// parseLevel преобразует строковое представление уровня в Level
func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Debug выводит отладочное сообщение
func (l *Logger) Debug(args ...interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", args...)
	}
}

// Info выводит информационное сообщение
func (l *Logger) Info(args ...interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", args...)
	}
}

// Warn выводит предупреждающее сообщение
func (l *Logger) Warn(args ...interface{}) {
	if l.level <= WarnLevel {
		l.log("WARN", args...)
	}
}

// Error выводит сообщение об ошибке
func (l *Logger) Error(args ...interface{}) {
	if l.level <= ErrorLevel {
		l.log("ERROR", args...)
	}
}

// log форматирует и выводит сообщение
func (l *Logger) log(level string, args ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)
	l.logger.Printf("[%s] %s: %s", now, level, message)
}
