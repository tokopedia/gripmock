package domain

import healthv1 "google.golang.org/grpc/health/grpc_health_v1"

type ServingStatus uint32

const (
	Unknown        = ServingStatus(healthv1.HealthCheckResponse_UNKNOWN)
	Serving        = ServingStatus(healthv1.HealthCheckResponse_SERVING)
	NotServing     = ServingStatus(healthv1.HealthCheckResponse_NOT_SERVING)
	ServiceUnknown = ServingStatus(healthv1.HealthCheckResponse_SERVICE_UNKNOWN)
)
