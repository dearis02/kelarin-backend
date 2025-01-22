package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kelarin/internal/types"
	"os"
	"sync"

	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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
	zerolog.TimeFieldFormat = types.TIME_FORMAT_TZ
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		if err, ok := err.(*errors.Error); ok {
			return FilterStackTrace(err.StackFrames())
		}

		return nil
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Logger = log.With().Stack().Logger()

	if c.PrettyLog {
		log.Logger = log.Output(jsonIndentWriter{Out: os.Stderr, TimeFormat: types.TIME_FORMAT_TZ})
		logger = logger.Output(jsonIndentWriter{Out: os.Stderr, TimeFormat: types.TIME_FORMAT_TZ})
	}

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
