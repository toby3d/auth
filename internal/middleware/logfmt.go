package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type (
	LogFmtConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Output is a writer where logs in JSON format are written.
		// Optional. Default value os.Stdout.
		Output io.Writer

		// TODO(toby3d): allow select some tags
	}

	logFmtResponse struct {
		http.ResponseWriter
		error          error
		start          time.Time
		end            time.Time
		statusCode     int
		responseLength int
		id             uint64
	}
)

//nolint:gochecknoglobals // default configuration
var DefaultLogFmtConfig = LogFmtConfig{
	Skipper: DefaultSkipper,
	Output:  os.Stdout,
}

//nolint:gochecknoglobals
var globalConnID uint64

func LogFmt() Interceptor {
	c := DefaultLogFmtConfig

	return LogFmtWithConfig(c)
}

func LogFmtWithConfig(config LogFmtConfig) Interceptor {
	if config.Skipper == nil {
		config.Skipper = DefaultLogFmtConfig.Skipper
	}

	if config.Output == nil {
		config.Output = DefaultLogFmtConfig.Output
	}

	encoder := slog.New(slog.NewTextHandler(config.Output, nil))

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		tx := &logFmtResponse{
			id:             nextConnID(),
			responseLength: 0,
			ResponseWriter: w,
			start:          time.Now().UTC(),
			statusCode:     0,
		}

		next(tx, r)

		tx.end = time.Now().UTC()
		payload := []any{
			slog.Int64("bytes_in", r.ContentLength),
			slog.Int("bytes_out", tx.responseLength),
			slog.Any("error", tx.error),
			slog.String("host", r.Host),
			slog.Uint64("id", tx.id),
			slog.Int64("latency", tx.end.Sub(tx.start).Nanoseconds()),
			slog.String("latency_human", tx.end.Sub(tx.start).String()),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("protocol", r.Proto),
			slog.String("referer", r.Referer()),
			slog.String("remote_ip", r.RemoteAddr),
			slog.Int("status", tx.statusCode),
			slog.String("time_rfc3339", tx.start.Format(time.RFC3339)),
			slog.String("time_rfc3339_nano", tx.start.Format(time.RFC3339Nano)),
			slog.Int64("time_unix", tx.start.Unix()),
			slog.Int64("time_unix_nano", tx.start.UnixNano()),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		}

		for name, src := range map[string]map[string][]string{
			"form":   r.PostForm,
			"header": r.Header,
			"query":  r.URL.Query(),
		} {
			values := make([]any, 0)

			for k, v := range src {
				values = append(values, slog.String(strings.ReplaceAll(strings.ToLower(k), " ", "_"),
					strings.Join(v, " ")))
			}

			payload = append(payload, slog.Group(name, values...))
		}

		encoder.Log(r.Context(), slog.LevelInfo, "" /* msg */, payload...)
	}
}

func (r *logFmtResponse) WriteHeader(status int) {
	r.statusCode = status

	r.ResponseWriter.WriteHeader(status)
}

func (r *logFmtResponse) Write(src []byte) (int, error) {
	var l int

	l, r.error = r.ResponseWriter.Write(src)
	r.responseLength += l

	return l, r.error
}

func nextConnID() uint64 {
	return atomic.AddUint64(&globalConnID, 1)
}
