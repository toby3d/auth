package middleware

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-logfmt/logfmt"
	http "github.com/valyala/fasthttp"
)

type LogFmtConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// TODO(toby3d): allow select some tags

	// Output is a writer where logs in JSON format are written.
	// Optional. Default value os.Stdout.
	Output io.Writer
}

var DefaultLogFmtConfig = LogFmtConfig{
	Skipper: DefaultSkipper,
	Output:  os.Stdout,
}

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

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		next(ctx)

		encoder.EncodeKeyvals(
			"bytes_in", len(ctx.Request.Body()),
			"bytes_out", len(ctx.Response.Body()),
			"error", ctx.Err(),
			"host", ctx.Host(),
			"id", ctx.ID(),
			"latency", ctx.Time().Sub(ctx.ConnTime()).Nanoseconds(),
			"latency_human", ctx.Time().Sub(ctx.ConnTime()).String(),
			"method", ctx.Method(),
			"path", ctx.Path(),
			"protocol", ctx.Request.Header.Protocol(),
			"referer", ctx.Referer(),
			"remote_ip", ctx.RemoteIP(),
			"status", ctx.Response.StatusCode(),
			"time_rfc3339", ctx.Time().Format(time.RFC3339),
			"time_rfc3339_nano", ctx.Time().Format(time.RFC3339Nano),
			"time_unix", ctx.Time().Unix(),
			"time_unix_nano", ctx.Time().UnixNano(),
			"uri", ctx.URI(),
			"user_agent", ctx.UserAgent(),
		)
		ctx.Request.Header.VisitAllInOrder(func(key, value []byte) {
			encoder.EncodeKeyval(strings.ReplaceAll(strings.ToLower("header_"+string(key)), "-", "_"), value)
		})
		ctx.QueryArgs().VisitAll(func(key, value []byte) {
			encoder.EncodeKeyval(strings.ReplaceAll(strings.ToLower("query_"+string(key)), "-", "_"), value)
		})

		if form, err := ctx.MultipartForm(); err == nil {
			for k, v := range form.Value {
				encoder.EncodeKeyval(strings.ReplaceAll(strings.ToLower("form_"+k), "-", "_"), v)
			}
		}

		encoder.EndRecord()
	}
}
