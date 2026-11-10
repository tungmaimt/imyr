package core

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

type AfterHandler = func(context.Context, fiber.Map)

// Define ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

// LogLevel defines our custom types
type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func logLvl(level LogLevel, s string, v ...any) {
	var color string
	switch level {
	case LevelInfo:
		color = Green
	case LevelWarn:
		color = Yellow
	case LevelError:
		color = Red
	default:
		color = Reset
	}

	msg := fmt.Sprintf(s, v...)
	logger.Printf("%s[%s]%s %s", color, level, Reset, msg)
}

func LogReq(c *fiber.Ctx) error {
	logLvl(LevelInfo, c.OriginalURL())

	body := string(c.Body())
	if len(body) > 0 {
		logLvl(LevelInfo, body)
	}

	return c.Next()
}

func LogURL(c *fiber.Ctx) error {
	logLvl(LevelInfo, "%s -- %s -- %s%s", c.Method(), c.IP(), c.BaseURL(), c.OriginalURL())
	return c.Next()
}
