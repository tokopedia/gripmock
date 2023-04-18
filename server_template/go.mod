module gripmock/generated

// This go.mod is used to pre-cache the modules required for the gripmock server
// builds in the dockerfile. It does not have to match the current requirements
// of the specific server being created, as "go mod tidy" will clean it up
// anyway, but it helps provide a cache and starting point.

go 1.19

require (
	cloud.google.com/go v0.105.0
	github.com/go-logr/stdr v1.2.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.40.0
	go.opentelemetry.io/contrib/propagators/autoprop v0.40.0
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.14.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.14.0
	go.opentelemetry.io/otel/exporters/zipkin v1.14.0
	go.opentelemetry.io/otel/sdk v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	golang.org/x/net v0.9.0
	google.golang.org/grpc v1.54.0
	google.golang.org/protobuf v1.30.0
	github.com/google/go-cmp v0.5.9
	github.com/stretchr/testify v1.8.2
	go.uber.org/goleak v1.2.1
	golang.org/x/oauth2 v0.4.0
	github.com/golang/glog v1.0.0
	gopkg.in/yaml.v3 v3.0.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/davecgh/go-spew v1.1.1
	google.golang.org/appengine v1.6.7
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/compute v1.15.1
	github.com/cenkalti/backoff/v4 v4.2.0
	github.com/go-logr/logr v1.2.3
	github.com/golang/protobuf v1.5.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/openzipkin/zipkin-go v0.4.1
	go.opentelemetry.io/contrib/propagators/aws v1.15.0
	go.opentelemetry.io/contrib/propagators/b3 v1.15.0
	go.opentelemetry.io/contrib/propagators/jaeger v1.15.0
	go.opentelemetry.io/contrib/propagators/ot v1.15.0
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.14.0
	go.opentelemetry.io/otel/metric v0.37.0
	go.opentelemetry.io/proto/otlp v0.19.0
	go.uber.org/atomic v1.7.0
	go.uber.org/multierr v1.9.0
	golang.org/x/sys v0.7.0
	golang.org/x/text v0.9.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
)
