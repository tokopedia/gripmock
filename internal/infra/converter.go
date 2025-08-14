package infra

import (
	"encoding/base64"
	"fmt"

	"github.com/goccy/go-json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ScalarConverter provides unified conversion for protobuf scalar values.
type ScalarConverter struct{}

// NewScalarConverter creates a new converter instance.
func NewScalarConverter() *ScalarConverter {
	return &ScalarConverter{}
}

// ConvertScalar converts a protobuf field value to its Go representation.
func (c *ScalarConverter) ConvertScalar(fd protoreflect.FieldDescriptor, value protoreflect.Value) any {
	const nullValue = "google.protobuf.NullValue"

	// Handle special cases first
	//nolint:exhaustive
	switch fd.Kind() {
	case protoreflect.EnumKind:
		return c.handleEnum(fd, value, nullValue)
	case protoreflect.MessageKind:
		return c.handleMessage(value)
	case protoreflect.GroupKind:
		return fmt.Sprintf("group type: %v", fd.Kind())
	}

	// Use map-based approach for scalar types
	handlers := c.scalarHandlers()
	if handler, ok := handlers[fd.Kind()]; ok {
		return handler(value)
	}

	return fmt.Sprintf("unknown type: %v", fd.Kind())
}

// ConvertMessage converts a protobuf message to a map representation.
// This method should be implemented by the caller or extended as needed.
func (c *ScalarConverter) ConvertMessage(msg proto.Message) map[string]any {
	if msg == nil {
		return nil
	}

	result := make(map[string]any)
	message := msg.ProtoReflect()

	message.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if !message.Has(fd) {
			return true
		}

		fieldName := string(fd.Name())

		// Handle different field types
		switch {
		case fd.IsList():
			result[fieldName] = c.convertList(fd, value.List())
		case fd.IsMap():
			result[fieldName] = c.convertMap(fd, value.Map())
		default:
			result[fieldName] = c.ConvertScalar(fd, value)
		}

		return true
	})

	return result
}

// convertList converts a protobuf list to a Go slice.
func (c *ScalarConverter) convertList(fd protoreflect.FieldDescriptor, list protoreflect.List) []any {
	result := make([]any, list.Len())

	for i := range list.Len() {
		result[i] = c.ConvertScalar(fd, list.Get(i))
	}

	return result
}

// convertMap converts a protobuf map to a Go map.
func (c *ScalarConverter) convertMap(fd protoreflect.FieldDescriptor, mapVal protoreflect.Map) map[string]any {
	result := make(map[string]any)

	mapVal.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		keyStr := key.String()
		result[keyStr] = c.ConvertScalar(fd.MapValue(), value)

		return true
	})

	return result
}

// scalarHandlers maps protobuf field kinds to their conversion functions.
func (c *ScalarConverter) scalarHandlers() map[protoreflect.Kind]func(protoreflect.Value) any {
	return map[protoreflect.Kind]func(protoreflect.Value) any{
		protoreflect.BoolKind: func(value protoreflect.Value) any {
			return value.Bool()
		},
		protoreflect.Int32Kind:    c.handleNumber,
		protoreflect.Sint32Kind:   c.handleNumber,
		protoreflect.Sfixed32Kind: c.handleNumber,
		protoreflect.Int64Kind:    c.handleNumber,
		protoreflect.Sint64Kind:   c.handleNumber,
		protoreflect.Sfixed64Kind: c.handleNumber,
		protoreflect.Uint32Kind:   c.handleNumber,
		protoreflect.Fixed32Kind:  c.handleNumber,
		protoreflect.Uint64Kind:   c.handleNumber,
		protoreflect.Fixed64Kind:  c.handleNumber,
		protoreflect.FloatKind:    c.handleNumber,
		protoreflect.DoubleKind:   c.handleNumber,
		protoreflect.StringKind: func(value protoreflect.Value) any {
			return value.String()
		},
		protoreflect.BytesKind: func(value protoreflect.Value) any {
			return base64.StdEncoding.EncodeToString(value.Bytes())
		},
	}
}

// handleNumber handles all numeric kinds and returns a json.Number.
func (c *ScalarConverter) handleNumber(value protoreflect.Value) any {
	return json.Number(value.String())
}

// handleEnum handles EnumKind fields, including google.protobuf.NullValue.
func (c *ScalarConverter) handleEnum(fd protoreflect.FieldDescriptor, value protoreflect.Value, nullValue string) any {
	if string(fd.Enum().FullName()) == nullValue {
		return nil
	}

	// Get the enum descriptor for the value
	desc := fd.Enum().Values().ByNumber(value.Enum())
	if desc != nil {
		return string(desc.Name())
	}

	return value.Enum()
}

// handleMessage handles MessageKind fields.
func (c *ScalarConverter) handleMessage(value protoreflect.Value) any {
	if value.Message().IsValid() {
		return c.ConvertMessage(value.Message().Interface())
	}

	return nil
}
