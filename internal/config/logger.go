package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeStampFormat = "2006-01-02 15:04:05 -0700"

var jsonIndentWriterBufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 100))
	},
}

type jsonIndentWriter struct {
	Out        io.Writer
	TimeFormat string
}

func (w jsonIndentWriter) Write(p []byte) (int, error) {
	var buf = jsonIndentWriterBufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		jsonIndentWriterBufPool.Put(buf)
	}()

	err := json.Indent(buf, p, "", "  ")
	if err != nil {
		return 0, err
	}

	_, err = buf.WriteTo(w.Out)
	return len(p), err
}

func NewLogger(c *Config) zerolog.Logger {
	zerolog.TimeFieldFormat = timeStampFormat
	zerolog.ErrorStackMarshaler = func(err error) any {
		if err, ok := err.(*errors.Error); ok {
			return FilterStackTrace(err.StackFrames())
		}

		return nil
	}

	logContext := log.With().Stack()
	logAppNameDict := zerolog.Dict().Str("name", "kelarin-backend")

	if c.Environment != "production" {
		log.Logger = logContext.Logger().Output(jsonIndentWriter{Out: os.Stderr, TimeFormat: timeStampFormat})
		return logContext.Logger().Output(jsonIndentWriter{Out: os.Stderr, TimeFormat: timeStampFormat})
	}

	logger := logContext.Dict("app", logAppNameDict).Logger()
	log.Logger = logContext.Dict("app", logAppNameDict).Logger()

	return logger
}

func FilterStackTrace(st []errors.StackFrame) []string {
	var filteredStack []string

	for _, frame := range st {
		frameStr := fmt.Sprintf("%s:%d", frame.File, frame.LineNumber)
		filteredStack = append(filteredStack, frameStr)
	}
	return filteredStack
}
