package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var (
	currentLevel Level = LevelInfo
	mu           sync.Mutex
)

// SetLevel sets the global log level.
func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = LevelDebug
	case "info":
		currentLevel = LevelInfo
	case "warn", "warning":
		currentLevel = LevelWarn
	case "error":
		currentLevel = LevelError
	}
}

// GetLevel returns the current log level name.
func GetLevel() string {
	return levelNames[currentLevel]
}

func logf(level Level, format string, args ...interface{}) {
	if level < currentLevel {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	ts := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] ", ts, levelNames[level])
	fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
}

func Debug(format string, args ...interface{}) { logf(LevelDebug, format, args...) }
func Info(format string, args ...interface{})  { logf(LevelInfo, format, args...) }
func Warn(format string, args ...interface{})  { logf(LevelWarn, format, args...) }
func Error(format string, args ...interface{}) { logf(LevelError, format, args...) }
