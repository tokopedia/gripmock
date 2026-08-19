package stuber

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"

	"github.com/bavix/gripmock/v3/internal/infra/types"
)

const (
	SourceFile  = "file"
	SourceRest  = "rest"
	SourceMCP   = "mcp"
	SourceProxy = "proxy"
)

// DelayType controls how the configured response delay changes between matches.
type DelayType string

const (
	// DelayTypeDefault keeps the configured delay unchanged for every match.
	DelayTypeDefault DelayType = "default"
	// DelayTypeRegressive decreases the configured delay after every match.
	DelayTypeRegressive DelayType = "regressive"
)

func IsKnownSource(source string) bool {
	switch source {
	case SourceFile, SourceRest, SourceMCP, SourceProxy:
		return true
	default:
		return false
	}
}

// StubOptions holds optional behavior settings for a stub.
type StubOptions struct {
	Times int `json:"times,omitempty" validate:"gte=0"` // Max number of matches; 0 = unlimited.
}

// Stub represents a gRPC service method and its associated data.
type Stub struct {
	ID                  uuid.UUID     `json:"id"`                                                           // Unique stub identifier.
	Service             string        `json:"service"                       validate:"required"`            // Service name.
	Method              string        `json:"method"                        validate:"required"`            // Method name.
	Session             string        `json:"session,omitempty"`                                            // Session ID (empty = global).
	Priority            int           `json:"priority"`                                                     // Priority score.
	Options             StubOptions   `json:"options,omitempty"`                                            //nolint:modernize
	Headers             InputHeader   `json:"headers"`                                                      // Request headers.
	Input               InputData     `json:"input"                         validate:"valid_input_config"`  // Unary matches.
	Inputs              []InputData   `json:"inputs,omitempty"              validate:"valid_input_config"`  // Stream matches.
	Output              Output        `json:"output"                        validate:"valid_output_config"` // Response output.
	Effects             []Effect      `json:"effects,omitempty"             validate:"valid_effects"`       // Side effects.
	Source              string        `json:"source,omitempty"`                                             // Stub source.
	Handler             StreamHandler `json:"-"`                                                            // Custom stream handler.
	MatchOnFirstMessage bool          `json:"matchOnFirstMessage,omitempty"`                                // Match on first msg only.
}

// StreamHandler processes a bidirectional stream directly.
// When set on a Stub, the handler manages send/recv instead of using static Output.
type StreamHandler func(ctx context.Context, stream any) error

const (
	EffectActionUpsert = "upsert"
	EffectActionDelete = "delete"
)

// Effect represents a side effect executed after stub match.
type Effect struct {
	Action string         `json:"action"`
	ID     string         `json:"id,omitempty"`
	Stub   map[string]any `json:"stub,omitempty"`
}

// GripMockKey is the reserved map key for per-element stream metadata.
// When present in a stream entry, it must be a map[string]any.
// Supported sub-keys:
//   - "delay" — per-element delay before sending this message (Go duration string)
const GripMockKey = "_gripmock"

// ExtractGripMockDelay checks a stream element map for GripMockKey.
// If found, it extracts the "delay" value, deletes GripMockKey from the map,
// and returns the parsed duration. Returns (0, false) if no valid delay is present.
//
// The map is modified in place (GripMockKey is removed).
func ExtractGripMockDelay(m map[string]any) (types.Duration, bool) {
	if m == nil {
		return 0, false
	}

	raw, has := m[GripMockKey]
	if !has {
		return 0, false
	}

	delete(m, GripMockKey)

	gk, ok := raw.(map[string]any)
	if !ok {
		return 0, false
	}

	delayStr, has := gk["delay"]
	if !has {
		return 0, false
	}

	s, ok := delayStr.(string)
	if !ok {
		return 0, false
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, false
	}

	return types.Duration(d), true
}

// EffectiveTimes returns the stub's max match count; 0 means unlimited.
func (s *Stub) EffectiveTimes() int {
	return s.Options.Times
}

// IsUnary returns true if this stub is for unary requests (has Input data).
func (s *Stub) IsUnary() bool {
	return len(s.Inputs) == 0
}

// IsClientStream returns true if this stub is for client streaming requests (has Inputs data).
func (s *Stub) IsClientStream() bool {
	return len(s.Inputs) > 0
}

// IsServerStream returns true if this stub is for server streaming responses (has Output.Stream data).
func (s *Stub) IsServerStream() bool {
	return len(s.Output.Stream) > 0
}

// IsBidirectional returns true if this stub can handle bidirectional streaming.
// For bidirectional streaming, the stub should have Inputs data (for input matching) and Output.Stream data (for output).
func (s *Stub) IsBidirectional() bool {
	return s.IsClientStream() && s.IsServerStream()
}

// Key returns the unique identifier of the stub.
func (s *Stub) Key() uuid.UUID {
	return s.ID
}

// Left returns the service name of the stub.
func (s *Stub) Left() string {
	return s.Service
}

// Right returns the method name of the stub.
func (s *Stub) Right() string {
	return s.Method
}

// Score returns the priority score of the stub.
func (s *Stub) Score() int {
	return s.Priority
}

// InputData represents the input data of a gRPC request.
type InputData struct {
	IgnoreArrayOrder bool           `json:"ignoreArrayOrder,omitempty"` // Whether to ignore the order of arrays in the input data.
	Equals           map[string]any `json:"equals"`                     // The data to match exactly.
	Contains         map[string]any `json:"contains"`                   // The data to match partially.
	Matches          map[string]any `json:"matches"`                    // The data to match using regular expressions.
	Glob             map[string]any `json:"glob,omitempty"`             // The data to match using glob patterns.
	AnyOf            []AnyOfElement `json:"anyOf,omitempty"`            // Alternative matchers (OR logic).
}

// AnyOfElement is a flat alternative matcher inside InputData.
// Unlike InputData, it cannot nest further AnyOf — depth is limited to 1.
type AnyOfElement struct {
	IgnoreArrayOrder bool           `json:"ignoreArrayOrder,omitempty"`
	Equals           map[string]any `json:"equals"`
	Contains         map[string]any `json:"contains"`
	Matches          map[string]any `json:"matches"`
	Glob             map[string]any `json:"glob,omitempty"`
}

// GetEquals returns the data to match exactly.
func (i InputData) GetEquals() map[string]any {
	return i.Equals
}

// GetContains returns the data to match partially.
func (i InputData) GetContains() map[string]any {
	return i.Contains
}

// GetMatches returns the data to match using regular expressions.
func (i InputData) GetMatches() map[string]any {
	return i.Matches
}

// GetGlob returns the data to match using glob patterns.
func (i InputData) GetGlob() map[string]any {
	return i.Glob
}

// InputHeader represents the headers of a gRPC request.
type InputHeader struct {
	Equals   map[string]any       `json:"equals"`          // The headers to match exactly.
	Contains map[string]any       `json:"contains"`        // The headers to match partially.
	Matches  map[string]any       `json:"matches"`         // The headers to match using regular expressions.
	Glob     map[string]any       `json:"glob,omitempty"`  // The headers to match using glob patterns.
	AnyOf    []AnyOfHeaderElement `json:"anyOf,omitempty"` // Alternative matchers (OR logic).
}

// AnyOfHeaderElement is a flat alternative matcher inside InputHeader.
type AnyOfHeaderElement struct {
	Equals   map[string]any `json:"equals"`
	Contains map[string]any `json:"contains"`
	Matches  map[string]any `json:"matches"`
	Glob     map[string]any `json:"glob,omitempty"`
}

// GetEquals returns the headers to match exactly.
func (i InputHeader) GetEquals() map[string]any {
	return i.Equals
}

// GetContains returns the headers to match partially.
func (i InputHeader) GetContains() map[string]any {
	return i.Contains
}

// GetMatches returns the headers to match using regular expressions.
func (i InputHeader) GetMatches() map[string]any {
	return i.Matches
}

// GetGlob returns the headers to match using glob patterns.
func (i InputHeader) GetGlob() map[string]any {
	return i.Glob
}

// Len returns the total number of headers to match.
func (i InputHeader) Len() int {
	n := len(i.Equals) + len(i.Matches) + len(i.Contains) + len(i.Glob)

	for _, alt := range i.AnyOf {
		n += len(alt.Equals) + len(alt.Matches) + len(alt.Contains) + len(alt.Glob)
	}

	return n
}

// Output represents the output data of a gRPC response.
type Output struct {
	Headers map[string]string `json:"headers"`          // The headers of the response.
	Data    any               `json:"data,omitempty"`   // The data of the response. Map for regular messages, scalar for WKT top-level.
	Stream  []any             `json:"stream,omitempty"` // The stream data for server-side streaming.
	// Each element represents a message to be sent. Each entry may be a map (regular message) or scalar (WKT top-level).
	Error   string           `json:"error"`             // The error message of the response.
	Code    *codes.Code      `json:"code,omitempty"`    // The status code of the response.
	Details []map[string]any `json:"details,omitempty"` // gRPC error details (google.protobuf.Any payloads).
	Delay   types.Duration   `json:"delay,omitempty"`   // The base delay of the response or error.
	//nolint:tagliatelle // Keep the proposed public stub field name.
	DelayType DelayType `json:"delay_type,omitempty"` // The strategy used to adjust Delay between matches.
	//nolint:tagliatelle // Keep the proposed public stub field name.
	DelayStep types.Duration `json:"delay_step,omitempty"` // The amount subtracted after each match.
}

// EffectiveDelay returns the delay for the one-based match number.
func (o Output) EffectiveDelay(matchNumber int) types.Duration {
	if o.DelayType != DelayTypeRegressive || matchNumber <= 1 || o.Delay <= 0 || o.DelayStep <= 0 {
		return o.Delay
	}

	decrements := int64(matchNumber - 1)
	delay := int64(o.Delay)
	step := int64(o.DelayStep)

	// Avoid overflowing the multiplication and clamp the result at zero.
	if decrements > delay/step {
		return 0
	}

	return types.Duration(delay - decrements*step)
}

func effectiveDelayStub(stub *Stub, matchNumber int) *Stub {
	if stub == nil || stub.Output.DelayType != DelayTypeRegressive || matchNumber <= 0 {
		return stub
	}

	result := *stub
	result.Output = stub.Output
	result.Output.Delay = stub.Output.EffectiveDelay(matchNumber)

	return &result
}
