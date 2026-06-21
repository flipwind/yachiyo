package logger

import (
	"fmt"
	"time"
)

type Logger struct {
	sourcename string
}

func New(sourcename string) *Logger{
	return &Logger{
		sourcename: sourcename,
	}
}

func (l *Logger) Success(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[32m[%s] [SUCCESS] [%s] %s\033[0m\n", timeStr, l.sourcename, msg)
}

func (l *Logger) Info(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[36m[%s] [INFO] [%s] %s\033[0m\n", timeStr, l.sourcename, msg)
}

func (l *Logger) Error(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[31m[%s] [ERROR] [%s] %s\033[0m\n", timeStr, l.sourcename, msg)
}

func (l *Logger) Warn(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[35m[%s] [WARN] [%s] %s\033[0m\n", timeStr, l.sourcename, msg)
}

func (l *Logger) Debug(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[34m[%s] [DEBUG] [%s] %s\033[0m\n", timeStr, l.sourcename, msg)
}