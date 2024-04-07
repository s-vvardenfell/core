package core

import (
	"errors"
	"os"
)

var (
	ErrNoServiceName = errors.New("service name not set or empty string")
)

type CoreOpts struct {
	ServiceName    *string
	JaegerGrpcAddr *string
	JaegerHttpAddr *string
	MetricsAddr    *string
}

func InitConfig() CoreOpts {
	serviceName := CheckEnv("SERVICE_NAME", "")
	jaegerGrpcAddr := CheckEnv("JAEGER_GRPC_ADDR", "0.0.0.0:4317")
	jaegerHttpAddr := CheckEnv("JAEGER_HTTP_ADDR", "0.0.0.0:4318")
	metricsAddr := CheckEnv("METRICS_ADDR", "0.0.0.0:9101")

	return CoreOpts{
		ServiceName:    &serviceName,
		JaegerGrpcAddr: &jaegerGrpcAddr,
		JaegerHttpAddr: &jaegerHttpAddr,
		MetricsAddr:    &metricsAddr,
	}
}

func CheckEnv(variableName, defaultValue string) string {
	if cn, ok := os.LookupEnv(variableName); ok {
		return cn
	}

	return defaultValue
}
