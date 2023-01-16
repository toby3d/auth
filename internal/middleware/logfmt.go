package middleware

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-logfmt/logfmt"
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

	encoder := logfmt.NewEncoder(config.Output)

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw := &logFmtResponse{
			id:             nextConnID(),
			responseLength: 0,
			ResponseWriter: w,
			start:          time.Now().UTC(),
			statusCode:     0,
		}

		next(rw, r)

		end := time.Now().UTC()

		encoder.EncodeKeyvals(
			"bytes_in", r.ContentLength,
			"bytes_out", rw.responseLength,
			"error", rw.error,
			"host", r.Host,
			"id", rw.id,
			"latency", end.Sub(rw.start).Nanoseconds(),
			"latency_human", end.Sub(rw.start).String(),
			"method", r.Method,
			"path", r.URL.Path,
			"protocol", r.Proto,
			"referer", r.Referer(),
			"remote_ip", r.RemoteAddr,
			"status", rw.statusCode,
			"time_rfc3339", rw.start.Format(time.RFC3339),
			"time_rfc3339_nano", rw.start.Format(time.RFC3339Nano),
			"time_unix", rw.start.Unix(),
			"time_unix_nano", rw.start.UnixNano(),
			"uri", r.RequestURI,
			"user_agent", r.UserAgent(),
		)

		for name, src := range map[string]map[string][]string{
			"form":   r.PostForm,
			"header": r.Header,
			"query":  r.URL.Query(),
		} {
			for k, v := range src {
				encoder.EncodeKeyval(name+"_"+strings.ReplaceAll(strings.ToLower(k), "-", "_"), v)
			}
		}

		encoder.EndRecord()
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
