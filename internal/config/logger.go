package config

import (
	"fmt"
	"kelarin/internal/types"
	"os"
	"strings"

	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger(c *Config) zerolog.Logger {
	zerolog.TimeFieldFormat = types.TIME_FORMAT_TZ
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		if err, ok := err.(*errors.Error); ok {
			endKeyword := "gin-gonic/gin"
			return FilterStackTrace(err.StackFrames(), endKeyword)
		}

		return nil
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Logger = log.With().Stack().Logger()
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: types.TIME_FORMAT_TZ}

	if c.PrettyLog {
		log.Logger = log.Output(consoleWriter)
		logger = logger.Output(consoleWriter)
	}

	return logger
}

func FilterStackTrace(st []errors.StackFrame, endKeyword string) []string {
	var filteredStack []string

	for _, frame := range st {
		frameStr := fmt.Sprintf("%s:%d", frame.File, frame.LineNumber)
		if endKeyword != "" && strings.Contains(frameStr, endKeyword) {
			break
		} else {
			filteredStack = append(filteredStack, frameStr)
		}
	}
	return filteredStack
}
