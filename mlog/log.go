package mlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type Level uint8

const (
	// PanicLevel 0
	PanicLevel Level = iota

	// FatalLevel 1
	FatalLevel

	// ErrorLevel 2
	ErrorLevel

	// WarnLevel 3
	WarnLevel

	// InfoLevel 4
	InfoLevel

	// DebugLevel 5
	DebugLevel
)

var (
	// DebugPrefix allows you to change its styling
	DebugPrefix = "[\033[32mDEBUG\033[0m]"

	// InfoPrefix allows you to change its styling
	InfoPrefix = "[\033[34m~INFO\033[0m]"

	// WarnPrefix allows you to change its styling
	WarnPrefix = "[\033[33m!WARN\033[0m]"

	// ErrorPrefix allows you to change its styling
	ErrorPrefix = "[\033[31mERROR\033[0m]"

	// FatalPrefix allows you to change its styling
	FatalPrefix = "[\033[31mFATAL\033[0m]"

	// PanicPrefix allows you to change its styling
	PanicPrefix = "[\033[31mPANIC\033[0m]"
)

// Logger holds logging configurations
type Logger struct {
	Level      Level
	Out        io.Writer
	Prefix     string
	Time       bool
	TimeFormat string
}

// NewLogger will initialize a new Logger struct
func New() *Logger {
	return &Logger{
		Out:        os.Stdout,
		Level:      InfoLevel,
		Prefix:     "",
		Time:       false,
		TimeFormat: "15:04:05",
	}
}

// SetLevel allows you to have the current log level
func (log *Logger) SetLevel(l Level) {
	log.Level = l
}

func (log *Logger) write(format string) {
	data := bytes.NewBuffer([]byte(format))

	_, err := io.Copy(log.Out, data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

func (log *Logger) output(prefix string, args ...interface{}) {
	log.write(fmt.Sprintf("%s %s %s %s\n",
		log.Prefix,
		time.Now().Format(log.TimeFormat),
		prefix,
		fmt.Sprint(args...),
	))
}

func (log *Logger) outputf(prefix, format string, args ...interface{}) {
	log.write(fmt.Sprintf("%s %s %s %s\n",
		log.Prefix,
		time.Now().Format(log.TimeFormat),
		prefix,
		fmt.Sprintf(format, args...),
	))
}

// Debug output
func (log *Logger) Debug(args ...interface{}) {
	if log.Level >= DebugLevel {
		log.output(DebugPrefix, args...)
	}
}

// Debugf output
func (log *Logger) Debugf(format string, args ...interface{}) {
	if log.Level >= DebugLevel {
		log.outputf(DebugPrefix, format, args...)
	}
}

// Info output
func (log *Logger) Info(args ...interface{}) {
	if log.Level >= InfoLevel {
		log.output(InfoPrefix, args...)
	}
}

// Infof output
func (log *Logger) Infof(format string, args ...interface{}) {
	if log.Level >= InfoLevel {
		log.outputf(InfoPrefix, format, args...)
	}
}

// Warn output
func (log *Logger) Warn(args ...interface{}) {
	if log.Level >= WarnLevel {
		log.output(WarnPrefix, args...)
	}
}

// Warnf output
func (log *Logger) Warnf(format string, args ...interface{}) {
	if log.Level >= WarnLevel {
		log.outputf(WarnPrefix, format, args...)
	}
}

// Error output
func (log *Logger) Error(args ...interface{}) {
	if log.Level >= ErrorLevel {
		log.output(ErrorPrefix, args...)
	}
}

// Errorf output
func (log *Logger) Errorf(format string, args ...interface{}) {
	if log.Level >= ErrorLevel {
		log.outputf(ErrorPrefix, format, args...)
	}
}

// Fatal output
func (log *Logger) Fatal(args ...interface{}) {
	if log.Level >= FatalLevel {
		log.output(FatalPrefix, args...)
		os.Exit(-1)
	}
}

// Fatalf output
func (log *Logger) Fatalf(format string, args ...interface{}) {
	if log.Level >= FatalLevel {
		log.outputf(FatalPrefix, format, args...)
		os.Exit(-1)
	}
}

// Panic output
func (log *Logger) Panic(args ...interface{}) {
	if log.Level >= PanicLevel {
		log.output(PanicPrefix, args...)
		os.Exit(-1)
	}
}

// Panicf output
func (log *Logger) Panicf(format string, args ...interface{}) {
	if log.Level >= PanicLevel {
		log.outputf(PanicPrefix, format, args...)
		os.Exit(-1)
	}
}
