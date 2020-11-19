package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	logger "github.com/rs/zerolog/log"
	"github.com/yusufsyaifudin/go-opentracing-example/tracer"
)

const (
	tracerURL   = "localhost:1111"
	serviceName = "RAINCOAT-EXTERNAL-SERVICE"
)

func main() {
	tracerService, closer := tracer.New(serviceName, tracerURL)
	defer closer.Close()

	// set global tracer of this application
	opentracing.SetGlobalTracer(tracerService)

	// Echo instance
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(traceMiddleware())

	e.GET("/get-the-raincoat", getTheRaincoat)

	// Start server
	logger.Info().Msg("starting server in :1324")
	err := e.Start(":1324")
	if err != nil {
		logger.Error().Err(err).Msg("")
		return
	}
}

// traceMiddleware add tracing into request context
func traceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()

			// set childCtx so each API request will creates new serverSpan log
			spanCtx, _ := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header),
			)

			serverSpan := opentracing.StartSpan(c.Request().URL.Path, ext.RPCServerOption(spanCtx))
			c.Set("serverSpan", serverSpan)

			defer func() {
				serverSpan.Finish()
			}()

			var headers []log.Field
			for k, v := range req.Header {
				headers = append(headers, log.String(k, strings.Join(v, ", ")))
			}

			serverSpan.LogFields(
				headers...,
			)

			// inject to response header
			_ = opentracing.GlobalTracer().Inject(
				serverSpan.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Response().Header()),
			)

			serverSpan.SetTag("endpoint", req.RequestURI)
			serverSpan.SetTag("host", req.Host)
			serverSpan.SetTag("clientIP", c.RealIP())
			serverSpan.SetTag("http.status", res.Status)
			serverSpan.SetTag("userAgent", req.UserAgent())

			start := time.Now()
			// continue the request
			if errMiddleware := next(c); errMiddleware != nil {
				c.Error(errMiddleware)
				c.Response().Committed = true
				return errMiddleware
			}

			stop := time.Now()

			logger.Debug().
				Float64("duration", float64(stop.Sub(start).Nanoseconds())/float64(time.Millisecond)).
				Int("status", res.Status).
				Str("protocol", req.Proto).
				Str("endpoint", req.RequestURI).
				Str("host", req.Host).
				Str("clientIP", c.RealIP()).
				Str("method", req.Method).
				Interface("resp_header", c.Response().Header()).
				Msg("handle request")

			return
		}
	}
}

func getTheRaincoat(eCtx echo.Context) error {
	serverSpan := eCtx.Get("serverSpan").(opentracing.Span)
	ctx := opentracing.ContextWithSpan(context.Background(), serverSpan)
	defer ctx.Done()

	lookingForAvailability(ctx)
	preparingTheRaincoat(ctx)

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("you got your raincoat"),
	})
}

func lookingForAvailability(parent context.Context) {
	span, ctx := opentracing.StartSpanFromContext(parent, "lookingForAvailability")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(10 * time.Millisecond)
}

func preparingTheRaincoat(parent context.Context) {
	span, ctx := opentracing.StartSpanFromContext(parent, "preparingTheRaincoat")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(20 * time.Millisecond)
}
