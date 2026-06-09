package logger

import (
	"fmt"
	"time"
)

func Success(source string, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[32m[%s] [SUCCESS] [%s] %s\033[0m\n", timeStr, source, msg)
}

func Info(source string, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[36m[%s] [INFO] [%s] %s\033[0m\n", timeStr, source, msg)
}

func Error(source string, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[31m[%s] [ERROR] [%s] %s\033[0m\n", timeStr, source, msg)
}

func Warn(source string, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[35m[%s] [WARN] [%s] %s\033[0m\n", timeStr, source, msg)
}

func Debug(source string, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	fmt.Printf("\033[34m[%s] [DEBUG] [%s] %s\033[0m\n", timeStr, source, msg)
}