package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	LevelDebug Level = iota // 0
	LevelInfo               //1
	LevelError              // 2
	LevelFatal              // 3
	LevelOff                // 4
)

//nolint:exhaustive
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties"`
		Trace      string            `json:"trace"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var logLine []byte
	logLine, err := json.Marshal(aux)
	if err != nil {
		logLine = []byte(LevelError.String() + ":unable to parse log message:" + err.Error())
	}

	// Avoid race conditions between logs newlines
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(logLine, '\n'))
}

// Write satisfies io.Writer interface to log error with no additional properties
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}

func (l *Logger) LogDebug(message string, properties map[string]string) {
	l.print(LevelDebug, message, properties)
}

func (l *Logger) LogInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) LogError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) LogFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}
