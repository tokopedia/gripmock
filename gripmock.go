package gripmock

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bavix/features"
	"github.com/gripmock/stuber"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/Dmytro-Hladkykh/gripmock/internal/domain"
)

// Server represents a simplified gRPC mock server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	budgerigar *stuber.Budgerigar
	port       int
	protoFiles []string
	mu         sync.RWMutex
	running    bool
}

// NewServer creates a new simplified gRPC mock server
func NewServer(port int, protoFiles []string) (*Server, error) {
	if port <= 0 {
		return nil, fmt.Errorf("invalid port: %d", port)
	}

	budgerigar := stuber.NewBudgerigar(features.New())

	server := &Server{
		budgerigar: budgerigar,
		port:       port,
		protoFiles: protoFiles,
	}

	if err := server.loadProtos(protoFiles); err != nil {
		return nil, fmt.Errorf("failed to load proto files: %w", err)
	}

	return server, nil
}

// Start starts the gRPC server on the specified port
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running on port %d", s.port)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	s.listener = listener
	s.grpcServer = grpc.NewServer()

	if err := s.registerServices(ctx); err != nil {
		listener.Close()
		return fmt.Errorf("failed to register services: %w", err)
	}

	s.running = true

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			// Server stopped
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	s.running = false
}

// AddStub adds a stub to the server
func (s *Server) AddStub(stub *stuber.Stub) error {
	stubs := s.budgerigar.PutMany(stub)
	if len(stubs) == 0 {
		return fmt.Errorf("failed to add stub")
	}
	return nil
}

// ClearStubs removes all stubs from the server
func (s *Server) ClearStubs() {
	s.budgerigar.Clear()
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	return s.port
}

// IsRunning returns true if the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// WaitForReady waits for the server to be ready to accept connections
func (s *Server) WaitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server not ready within timeout")
		case <-ticker.C:
			if s.IsRunning() {
				// Try to connect to verify server is actually accepting connections
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", s.port), 100*time.Millisecond)
				if err == nil {
					conn.Close()
					return nil
				}
			}
		}
	}
}

func (s *Server) loadProtos(protoFiles []string) error {
	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files specified")
	}

	params := domain.New(protoFiles, []string{})

	// Load the proto descriptors
	_, err := domain.Build(context.Background(), params.Imports(), params.ProtoPath())
	if err != nil {
		return fmt.Errorf("failed to build proto descriptors: %w", err)
	}

	return nil
}

func (s *Server) registerServices(ctx context.Context) error {
	params := domain.New(s.protoFiles, []string{})

	descriptors, err := domain.Build(ctx, params.Imports(), params.ProtoPath())
	if err != nil {
		return fmt.Errorf("failed to build descriptors: %w", err)
	}

	for _, descriptor := range descriptors {
		for _, file := range descriptor.GetFile() {
			for _, svc := range file.GetService() {
				serviceDesc := s.createServiceDesc(file, svc)
				s.registerServiceMethods(&serviceDesc, svc)
				s.grpcServer.RegisterService(&serviceDesc, nil)
			}
		}
	}

	return nil
}

func (s *Server) createServiceDesc(file *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto) grpc.ServiceDesc {
	serviceName := svc.GetName()
	if file.GetPackage() != "" {
		serviceName = fmt.Sprintf("%s.%s", file.GetPackage(), svc.GetName())
	}

	return grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*interface{})(nil),
	}
}

func (s *Server) registerServiceMethods(serviceDesc *grpc.ServiceDesc, svc *descriptorpb.ServiceDescriptorProto) {
	for _, method := range svc.GetMethod() {
		mocker := &SimpleMocker{
			budgerigar:      s.budgerigar,
			fullServiceName: serviceDesc.ServiceName,
			methodName:      method.GetName(),
		}

		if method.GetServerStreaming() || method.GetClientStreaming() {
			serviceDesc.Streams = append(serviceDesc.Streams, grpc.StreamDesc{
				StreamName:    method.GetName(),
				Handler:       mocker.streamHandler,
				ServerStreams: method.GetServerStreaming(),
				ClientStreams: method.GetClientStreaming(),
			})
		} else {
			serviceDesc.Methods = append(serviceDesc.Methods, grpc.MethodDesc{
				MethodName: method.GetName(),
				Handler:    mocker.unaryHandler,
			})
		}
	}
}


// NewStub creates a new stub with the given service, method, input and output
// User will need to create stuber.Stub manually for now
// func NewStub(service, method string, input map[string]interface{}, output map[string]interface{}) *stuber.Stub {}