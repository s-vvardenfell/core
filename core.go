package core

import (
	"context"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type Core struct {
	Logger            *zerolog.Logger
	HttpTraceProvider trace.TracerProvider
	GrpcTraceProvider trace.TracerProvider
	PromReg           *prometheus.Registry
}

func NewCore(ctx context.Context, opts CoreOpts) (*Core, error) {
	var (
		logger            = zerolog.New(os.Stdout).With().Timestamp().Logger()
		err               error
		httpTraceProvider trace.TracerProvider
		grpcTraceProvider trace.TracerProvider
		promReg           *prometheus.Registry
	)

	if opts.ServiceName == nil || *opts.ServiceName == "" {
		return nil, ErrNoServiceName
	}

	if opts.JaegerHttpAddr != nil {
		httpTraceProvider, err = initHttpTracer(ctx, *opts.ServiceName, *opts.JaegerHttpAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init http-trace-provider")
		}

		logger.Info().Str("service", *opts.ServiceName).Msg("init http trace provider")
	}

	if opts.JaegerGrpcAddr != nil {
		grpcTraceProvider, err = initGrpcTracer(ctx, *opts.ServiceName, *opts.JaegerGrpcAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init grpc-trace-provider")
		}

		logger.Info().Str("service", *opts.ServiceName).Msg("init grpc trace provider")
	}

	// ----- metrics server -----

	if opts.MetricsAddr != nil {
		promReg = prometheus.NewRegistry()

		go func() {
			http.Handle("/metrics", promhttp.HandlerFor(promReg, promhttp.HandlerOpts{
				Registry: promReg,
				ErrorLog: &logger,
			}))

			logger.Info().Str("service", "MetricsServer").Msgf("listening at %s", *opts.MetricsAddr)

			if err := http.ListenAndServe(*opts.MetricsAddr, nil); err != nil {
				logger.Error().Err(err).Msg("failed to run metrics exporter endpoint")
			}
		}()
	}

	return &Core{
		Logger:            &logger,
		HttpTraceProvider: httpTraceProvider,
		GrpcTraceProvider: grpcTraceProvider,
		PromReg:           promReg,
	}, nil
}

// func (c *Core) Shutdown(ctx context.Context) error {
// 	if err := c.HttpTraceProvider.Shutdown(ctx); err != nil {
// 		logger.Fatal().Err(err).Msg("failed to shutting down tracer provider")
// 	}

// 	if err := c.GrpcTraceProvider.Shutdown(ctx); err != nil {
// 		logger.Fatal().Err(err).Msg("failed to shutting down tracer provider")
// 	}
// }
