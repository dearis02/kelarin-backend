package workerUtil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

type WorkerLogger struct {
	logger zerolog.Logger
}

var jsonIndentWriterBufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 100))
	},
}

type jsonIndentWriter struct {
	Out io.Writer
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

func NewWorkerLogger() *WorkerLogger {
	workerDict := zerolog.Dict().Str("app", "ms-group-backend")
	logger := zerolog.New(jsonIndentWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Stack().
		Dict("worker", workerDict).
		Logger()

	return &WorkerLogger{logger: logger}
}

func (z *WorkerLogger) Debug(args ...interface{}) {
	z.logger.Debug().Msgf("%v", args...)
}

func (z *WorkerLogger) Info(args ...interface{}) {
	z.logger.Info().Msgf("%v", args...)
}

func (z *WorkerLogger) Warn(args ...interface{}) {
	z.logger.Warn().Msgf("%v", args...)
}

func (z *WorkerLogger) Error(args ...interface{}) {
	z.logger.Error().Msgf("%v", args...)
}

func (z *WorkerLogger) Fatal(args ...interface{}) {
	z.logger.Fatal().Msgf("%v", args...)
}

type ErrorHandler struct {
	logger zerolog.Logger
}

func NewErrorHandler(workerLogger *WorkerLogger) ErrorHandler {
	return ErrorHandler{logger: workerLogger.logger}
}

func (u ErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	u.logger.Error().
		Str("task_id", task.ResultWriter().TaskID()).
		Str("type", task.Type()).
		Any("payload", task.Payload()).
		Err(err).Msgf("error processing task")
}
