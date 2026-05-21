package logger

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

var globalLogger *Logger

func init() {
	globalLogger = &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile),
		warn:  log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile),
		err:   log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile),
	}
}

func Info(msg string, args ...interface{}) {
	globalLogger.info.Printf(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	globalLogger.warn.Printf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	globalLogger.err.Printf(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	globalLogger.debug.Printf(msg, args...)
}

func Infof(format string, args ...interface{}) {
	globalLogger.info.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.warn.Printf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.err.Printf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.debug.Printf(format, args...)
}
