package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	logger "github.com/rs/zerolog/log"
	"github.com/uber/jaeger-client-go"
	"github.com/yusufsyaifudin/go-opentracing-example/tracer"
)

const (
	tracerURL   = "localhost:5775"
	serviceName = "API-SERVICE"
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

	e.GET("/dora-the-explorer", doExplore)

	// Start server
	logger.Info().Msg("starting server in :1323")
	err := e.Start(":1323")
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

			traceID := "no-tracer-id"
			if sc, ok := serverSpan.Context().(jaeger.SpanContext); ok {
				traceID = sc.String()
			}

			// inject to response header
			opentracing.GlobalTracer().Inject(
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
				Str("tracerID", traceID).
				Msg("handle request")

			return
		}
	}
}

func doExplore(eCtx echo.Context) error {
	serverSpan := eCtx.Get("serverSpan").(opentracing.Span)
	ctx := opentracing.ContextWithSpan(context.Background(), serverSpan)
	defer ctx.Done()

	var (
		isRainyDay = false
		in         = int64(0)
		out        int64
	)

	isRainyDayStr := strings.TrimSpace(strings.ToLower(eCtx.QueryParam("is_rainy_day")))
	if isRainyDayStr == "true" || isRainyDayStr == "1" {
		isRainyDay = true
	}

	out, _ = startTheJourneyFromTreeHouse(ctx, in)
	out, _ = passTheForest(ctx, out)
	out, _ = crossTheLake(ctx, out, isRainyDay)
	out, _ = enterThePyramid(ctx, out)

	sandCastleSpan, _ := opentracing.StartSpanFromContext(ctx, "arrivedInSandCastle")
	defer sandCastleSpan.Finish()

	// process in sand castle
	time.Sleep(3 * time.Millisecond)

	var message = "clear weather"
	if isRainyDay {
		message = "rainy day"
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("we arrived in sand castle on a %s", message),
	})
}

func startTheJourneyFromTreeHouse(parent context.Context, param int64) (out int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(parent, "startTheJourneyFromTreeHouse")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(2 * time.Millisecond)
	span.SetTag("message", "prepare some food")

	out = param + 2
	return
}

func passTheForest(parent context.Context, param int64) (out int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(parent, "passTheForest")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(5 * time.Millisecond)
	span.SetTag("message", "oh it's a long forest and I'm getting tired now!")

	out = param + 5
	return
}

func crossTheLake(parent context.Context, param int64, isRainyDay bool) (out int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(parent, "crossTheLake")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(10 * time.Millisecond)
	out = param + 10

	if isRainyDay {
		time.Sleep(20 * time.Millisecond)
		span.SetTag("message", "It's a rainy day and "+
			"I must extra careful since I don't want my boat drowned with me")
	} else {
		span.SetTag("message", "Clear weather and I enjoy the view from the lake!")
	}

	span.SetTag("isRainyDay", isRainyDay)

	return
}

func enterThePyramid(parent context.Context, param int64) (out int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(parent, "enterThePyramid")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	time.Sleep(10 * time.Millisecond)
	out = param + 10

	span.SetTag("message", "Whoa, everywhere's dark!")
	return
}
