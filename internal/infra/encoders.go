package infra

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/bavix/apis/pkg/uuidconv"
)

// Base64Utils provides utility functions for base64 encoding.
type Base64Utils struct{}

// NewBase64Utils creates a new Base64Utils instance.
func NewBase64Utils() *Base64Utils {
	return &Base64Utils{}
}

// StringToBase64 encodes a string to base64.
func (u *Base64Utils) StringToBase64(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

// BytesToBase64 encodes bytes to base64.
func (u *Base64Utils) BytesToBase64(v []byte) string {
	return base64.StdEncoding.EncodeToString(v)
}

// UUIDUtils provides utility functions for UUID operations.
type UUIDUtils struct{}

// NewUUIDUtils creates a new UUIDUtils instance.
func NewUUIDUtils() *UUIDUtils {
	return &UUIDUtils{}
}

// UUIDToBase64 converts a UUID string to base64.
func (u *UUIDUtils) UUIDToBase64(input string) string {
	v := uuid.MustParse(input)

	return base64.StdEncoding.EncodeToString(v[:])
}

// UUIDToBytes converts a UUID string to a byte slice.
func (u *UUIDUtils) UUIDToBytes(input string) []byte {
	v := uuid.MustParse(input)

	return v[:]
}

// UUIDToInt64 converts a UUID string to high/low int64 JSON representation.
func (u *UUIDUtils) UUIDToInt64(str string) string {
	v := uuid.MustParse(str)

	high, low := uuidconv.UUID2DoubleInt(v)

	var sb strings.Builder

	sb.Grow(32) //nolint:mnd
	sb.WriteString(`{"high":`)
	sb.WriteString(strconv.FormatInt(high, 10))
	sb.WriteString(`,"low":`)
	sb.WriteString(strconv.FormatInt(low, 10))
	sb.WriteString(`}`)

	return sb.String()
}

// ConversionUtils provides basic conversion utilities.
type ConversionUtils struct{}

// NewConversionUtils creates a new ConversionUtils instance.
func NewConversionUtils() *ConversionUtils {
	return &ConversionUtils{}
}

// StringToBytes converts a string to a byte slice.
func (u *ConversionUtils) StringToBytes(v string) []byte {
	return []byte(v)
}

// TemplateUtils provides all encoding utilities for template functions.
type TemplateUtils struct {
	Base64     *Base64Utils
	UUID       *UUIDUtils
	Conversion *ConversionUtils
}

// NewTemplateUtils creates a new TemplateUtils instance with all utilities.
func NewTemplateUtils() *TemplateUtils {
	return &TemplateUtils{
		Base64:     NewBase64Utils(),
		UUID:       NewUUIDUtils(),
		Conversion: NewConversionUtils(),
	}
}
