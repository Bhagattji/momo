package logging

import (
	"fmt"
	"io"
	"log"
	"os"
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

var (
	mu       sync.Mutex
	logger   *log.Logger
	curLevel Level
	enabled  bool
)

func Init(debug bool) {
	mu.Lock()
	if enabled {
		mu.Unlock()
		return
	}
	enabled = true
	if debug {
		curLevel = LevelDebug
	} else {
		curLevel = LevelInfo
	}
	logFile := os.Getenv("MOMO_LOG_FILE")
	var w io.Writer = os.Stderr
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err == nil {
			w = f
		}
	}
	logger = log.New(w, "", 0)
	lvl := levelString()
	mu.Unlock()
	logAtUnsafe(LevelInfo, "INFO", "logging initialized (level=%s)", lvl)
}

func Debug(format string, args ...any) {
	logAt(LevelDebug, "DEBUG", format, args...)
}

func Info(format string, args ...any) {
	logAt(LevelInfo, "INFO", format, args...)
}

func Warn(format string, args ...any) {
	logAt(LevelWarn, "WARN", format, args...)
}

func Error(format string, args ...any) {
	logAt(LevelError, "ERROR", format, args...)
}

func logAt(lvl Level, tag string, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	logAtUnsafe(lvl, tag, format, args...)
}

func logAtUnsafe(lvl Level, tag string, format string, args ...any) {
	if !enabled || lvl < curLevel || logger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	ts := time.Now().Format(time.RFC3339)
	logger.Printf("[%s] %s %s", tag, ts, msg)
}

func Shutdown() {
	Info("session ended")
}

func SetLevel(lvl Level) {
	mu.Lock()
	defer mu.Unlock()
	curLevel = lvl
}

func levelString() string {
	switch curLevel {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

func IsDebug() bool {
	mu.Lock()
	defer mu.Unlock()
	return curLevel == LevelDebug
}