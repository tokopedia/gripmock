package gripmock

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/goccy/go-json"
	"github.com/gripmock/stuber"
)

type SimpleMocker struct {
	budgerigar      *stuber.Budgerigar
	fullServiceName string
	methodName      string
}

func (m *SimpleMocker) unaryHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	// Get message descriptors
	inputDesc, outputDesc, err := m.getMessageDescriptors()
	if err != nil {
		return nil, err
	}

	req := dynamicpb.NewMessage(inputDesc)
	if err := dec(req); err != nil {
		return nil, err
	}

	query := stuber.Query{
		Service: m.fullServiceName,
		Method:  m.methodName,
		Data:    m.convertToMap(req),
	}

	// Add headers if present
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		query.Headers = m.processHeaders(md)
	}

	result, err := m.budgerigar.FindByQuery(query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find stub: %v", err)
	}

	found := result.Found()
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "no stub found for service %s, method %s", m.fullServiceName, m.methodName)
	}

	// Convert response to dynamic message
	outputMsg, err := m.newOutputMessage(found.Output.Data, outputDesc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create response: %v", err)
	}

	return outputMsg, nil
}

func (m *SimpleMocker) streamHandler(srv interface{}, stream grpc.ServerStream) error {
	return status.Errorf(codes.Unimplemented, "streaming not implemented in simplified version")
}

func (m *SimpleMocker) getMessageDescriptors() (protoreflect.MessageDescriptor, protoreflect.MessageDescriptor, error) {
	// This is simplified - in a real implementation we would need to properly
	// resolve message types from the service definition
	return nil, nil, fmt.Errorf("message descriptor resolution not implemented")
}

func (m *SimpleMocker) convertToMap(msg proto.Message) map[string]interface{} {
	if msg == nil {
		return nil
	}

	result := make(map[string]interface{})
	message := msg.ProtoReflect()

	message.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if !message.Has(fd) {
			return true
		}

		fieldName := string(fd.Name())
		result[fieldName] = m.convertValue(fd, value)
		return true
	})

	return result
}

func (m *SimpleMocker) convertValue(fd protoreflect.FieldDescriptor, value protoreflect.Value) interface{} {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return value.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32(value.Int())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return value.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(value.Uint())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return value.Uint()
	case protoreflect.FloatKind:
		return float32(value.Float())
	case protoreflect.DoubleKind:
		return value.Float()
	case protoreflect.StringKind:
		return value.String()
	case protoreflect.BytesKind:
		return value.Bytes()
	case protoreflect.EnumKind:
		return string(fd.Enum().Values().ByNumber(value.Enum()).Name())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return m.convertToMap(value.Message().Interface())
	default:
		return nil
	}
}

func (m *SimpleMocker) processHeaders(md metadata.MD) map[string]interface{} {
	if len(md) == 0 {
		return nil
	}

	headers := make(map[string]interface{})
	excludedHeaders := []string{":authority", "content-type", "grpc-accept-encoding", "user-agent", "accept-encoding"}

	for k, v := range md {
		skip := false
		for _, excluded := range excludedHeaders {
			if k == excluded {
				skip = true
				break
			}
		}
		if !skip {
			headers[k] = strings.Join(v, ";")
		}
	}

	return headers
}

func (m *SimpleMocker) newOutputMessage(data map[string]interface{}, outputDesc protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	msg := dynamicpb.NewMessage(outputDesc)
	err = protojson.Unmarshal(jsonData, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON into dynamic message: %w", err)
	}

	return msg, nil
}

func getMessageDescriptor(messageType string) (protoreflect.MessageDescriptor, error) {
	msgName := protoreflect.FullName(strings.TrimPrefix(messageType, "."))

	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(msgName)
	if err != nil {
		return nil, fmt.Errorf("message descriptor not found: %v", err)
	}

	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message descriptor: %s", msgName)
	}

	return msgDesc, nil
}