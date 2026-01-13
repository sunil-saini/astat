package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
)

var (
	debugColor   = color.New(color.FgHiCyan)
	infoColor    = color.New(color.FgHiBlue)
	successColor = color.New(color.FgHiGreen)
	warnColor    = color.New(color.FgYellow)
	errorColor   = color.New(color.FgHiRed, color.Bold)

	mu sync.Mutex
)

func log(level Level, msg string, a ...any) {
	mu.Lock()
	defer mu.Unlock()

	timestamp := time.Now().Format("15:04:05")
	var prefix string
	var c *color.Color

	switch level {
	case LevelDebug:
		prefix = "[DEBUG]"
		c = debugColor
	case LevelInfo:
		prefix = "[INFO]"
		c = infoColor
	case LevelSuccess:
		prefix = "[SUCCESS]"
		c = successColor
	case LevelWarn:
		prefix = "[WARN]"
		c = warnColor
	case LevelError:
		prefix = "[ERROR]"
		c = errorColor
	}

	fmt.Fprintf(os.Stderr, "%s ", color.New(color.FgHiBlack).Sprint(timestamp))
	c.Fprintf(os.Stderr, "%-9s", prefix)
	fmt.Fprintf(os.Stderr, " %s\n", fmt.Sprintf(msg, a...))
}

func Debug(msg string, a ...any)   { log(LevelDebug, msg, a...) }
func Info(msg string, a ...any)    { log(LevelInfo, msg, a...) }
func Success(msg string, a ...any) { log(LevelSuccess, msg, a...) }
func Warn(msg string, a ...any)    { log(LevelWarn, msg, a...) }
func Error(msg string, a ...any)   { log(LevelError, msg, a...) }
