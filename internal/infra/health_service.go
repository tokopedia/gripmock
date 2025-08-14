package infra

import (
	"context"
	"time"

	"google.golang.org/grpc"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/Dmytro-Hladkykh/gripmock/internal/domain"
)

type Service struct {
	client healthv1.HealthClient
}

func NewService(client healthv1.HealthClient) *Service {
	return &Service{client: client}
}

func (s *Service) PingWithTimeout(
	ctx context.Context,
	timeout time.Duration,
	service string,
) (domain.ServingStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.Ping(ctx, service)
}

func (s *Service) Ping(ctx context.Context, service string) (domain.ServingStatus, error) {
	check, err := s.client.Check(
		ctx,
		&healthv1.HealthCheckRequest{Service: service},
		grpc.WaitForReady(true),
	)
	if err != nil {
		return domain.Unknown, err //nolint:wrapcheck
	}

	switch check.GetStatus() {
	case healthv1.HealthCheckResponse_SERVING:
		return domain.Serving, nil
	case healthv1.HealthCheckResponse_NOT_SERVING:
		return domain.NotServing, nil
	case healthv1.HealthCheckResponse_SERVICE_UNKNOWN:
		return domain.ServiceUnknown, nil
	case healthv1.HealthCheckResponse_UNKNOWN:
		return domain.Unknown, nil
	default:
		return domain.Unknown, nil
	}
}
