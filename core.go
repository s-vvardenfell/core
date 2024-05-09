package core

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type HealthResponse struct {
	Status string `json:"status"`
}

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

	if opts.JaegerHttpAddr != nil || *opts.JaegerHttpAddr != "" {
		httpTraceProvider, err = initHttpTracer(ctx, *opts.ServiceName, *opts.JaegerHttpAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init http-trace-provider")
		}

		logger.Info().Str("service", *opts.ServiceName).Msg("init http trace provider")
	}

	if opts.JaegerGrpcAddr != nil || *opts.JaegerHttpAddr != "" {
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

	// ----- healthcheck server -----

	if opts.HealthCheckAddr != nil {
		go func() {
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(HealthResponse{Status: "ready"})
			})

			logger.Info().Str("service", "HealthcheckServer").Msgf("listening at %s", *opts.HealthCheckAddr)

			if err := http.ListenAndServe(*opts.HealthCheckAddr, nil); err != nil {
				logger.Error().Err(err).Msg("failed to run healthcheck server endpoint")
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
