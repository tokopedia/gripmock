package stuber_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/bavix/gripmock/v3/internal/infra/stuber"
	"github.com/bavix/gripmock/v3/internal/infra/types"
)

func TestStubMethods(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	stub := &stuber.Stub{
		ID:       id,
		Service:  "TestService",
		Method:   "TestMethod",
		Priority: 10,
	}

	require.Equal(t, id, stub.Key())
	require.Equal(t, "TestService", stub.Left())
	require.Equal(t, "TestMethod", stub.Right())
	require.Equal(t, 10, stub.Score())
}

func TestInputDataMethods(t *testing.T) {
	t.Parallel()

	inputData := stuber.InputData{
		IgnoreArrayOrder: true,
		Equals:           map[string]any{"key1": "value1"},
		Contains:         map[string]any{"key2": "value2"},
		Matches:          map[string]any{"key3": "value3"},
	}

	require.Equal(t, map[string]any{"key1": "value1"}, inputData.GetEquals())
	require.Equal(t, map[string]any{"key2": "value2"}, inputData.GetContains())
	require.Equal(t, map[string]any{"key3": "value3"}, inputData.GetMatches())
}

func TestInputHeaderMethods(t *testing.T) {
	t.Parallel()

	inputHeader := stuber.InputHeader{
		Equals:   map[string]any{"header1": "value1"},
		Contains: map[string]any{"header2": "value2"},
		Matches:  map[string]any{"header3": "value3"},
	}

	require.Equal(t, map[string]any{"header1": "value1"}, inputHeader.GetEquals())
	require.Equal(t, map[string]any{"header2": "value2"}, inputHeader.GetContains())
	require.Equal(t, map[string]any{"header3": "value3"}, inputHeader.GetMatches())
	require.Equal(t, 3, inputHeader.Len())
}

func TestInputHeaderLenEmpty(t *testing.T) {
	t.Parallel()

	inputHeader := stuber.InputHeader{}

	require.Equal(t, 0, inputHeader.Len())
}

func TestOutputFields(t *testing.T) {
	t.Parallel()

	code := codes.OK
	output := stuber.Output{
		Headers: map[string]string{"header1": "value1"},
		Data:    map[string]any{"data1": "value1"},
		Stream:  []any{"message1", "message2", "message3"},
		Error:   "test error",
		Code:    &code,
		Details: []map[string]any{{
			"type":   "type.googleapis.com/google.rpc.ErrorInfo",
			"reason": "UNIT_TEST",
		}},
		Delay: types.Duration(100),
	}

	require.Equal(t, map[string]string{"header1": "value1"}, output.Headers)
	require.Equal(t, map[string]any{"data1": "value1"}, output.Data)
	require.Equal(t, []any{"message1", "message2", "message3"}, output.Stream)
	require.Equal(t, "test error", output.Error)
	require.Equal(t, &code, output.Code)
	require.Len(t, output.Details, 1)
	require.Equal(t, 100, int(output.Delay))
}

func TestOutputFieldsEmptyStream(t *testing.T) {
	t.Parallel()

	output := stuber.Output{
		Headers: map[string]string{"header1": "value1"},
		Data:    map[string]any{"data1": "value1"},
		// Stream field is not set (should be nil)
	}

	require.Equal(t, map[string]string{"header1": "value1"}, output.Headers)
	require.Equal(t, map[string]any{"data1": "value1"}, output.Data)
	require.Nil(t, output.Stream)
}

func TestOutputFieldsOptionalDelay(t *testing.T) {
	t.Parallel()

	output := stuber.Output{
		Headers: map[string]string{"header1": "value1"},
		Data:    map[string]any{"data1": "value1"},
		// Delay field is not set (should be zero value)
	}

	require.Equal(t, map[string]string{"header1": "value1"}, output.Headers)
	require.Equal(t, map[string]any{"data1": "value1"}, output.Data)
	require.Equal(t, types.Duration(0), output.Delay)
}

func TestExtractGripMockDelay(t *testing.T) {
	t.Parallel()

	t.Run("no_gripmock_key", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"status": "OK"}
		d, ok := stuber.ExtractGripMockDelay(m)
		require.False(t, ok)
		require.Zero(t, d)
		require.Equal(t, map[string]any{"status": "OK"}, m)
	})

	t.Run("with_valid_delay", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"status": "OK", "_gripmock": map[string]any{"delay": "150ms"}}
		d, ok := stuber.ExtractGripMockDelay(m)
		require.True(t, ok)
		require.Equal(t, types.Duration(150*time.Millisecond), d)

		_, has := m["_gripmock"]
		require.False(t, has)
		require.Equal(t, map[string]any{"status": "OK"}, m)
	})

	t.Run("invalid_delay_string", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"_gripmock": map[string]any{"delay": "not-a-duration"}}
		d, ok := stuber.ExtractGripMockDelay(m)
		require.False(t, ok)
		require.Zero(t, d)
	})

	t.Run("nil_map", func(t *testing.T) {
		t.Parallel()

		d, ok := stuber.ExtractGripMockDelay(nil)
		require.False(t, ok)
		require.Zero(t, d)
	})

	t.Run("gripmock_not_a_map", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"_gripmock": "string-not-map"}
		d, ok := stuber.ExtractGripMockDelay(m)
		require.False(t, ok)
		require.Zero(t, d)

		_, has := m["_gripmock"]
		require.False(t, has)
	})
}

func TestOutputDelayJsonSerialization(t *testing.T) {
	t.Parallel()
	// Test JSON serialization with delay
	output := stuber.Output{
		Headers: map[string]string{"content-type": "application/json"},
		Data:    map[string]any{"message": "Hello World"},
		Delay:   types.Duration(100 * time.Millisecond),
	}

	jsonData, err := json.Marshal(output)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), `"delay":"100ms"`)

	// Test JSON deserialization with delay
	var decodedOutput stuber.Output

	err = json.Unmarshal([]byte(`{"headers":{"content-type":"application/json"},"data":{"message":"Hello World"},"delay":"200ms"}`),
		&decodedOutput)
	require.NoError(t, err)
	require.Equal(t, types.Duration(200*time.Millisecond), decodedOutput.Delay)
}

func TestOutputRegressiveDelayJSONSerialization(t *testing.T) {
	t.Parallel()

	output := stuber.Output{
		Data:      map[string]any{"message": "Hello World"},
		Delay:     types.Duration(3 * time.Second),
		DelayType: stuber.DelayTypeRegressive,
		DelayStep: types.Duration(500 * time.Millisecond),
	}

	jsonData, err := json.Marshal(output)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), `"delay":"3s"`)
	require.Contains(t, string(jsonData), `"delay_type":"regressive"`)
	require.Contains(t, string(jsonData), `"delay_step":"500ms"`)

	var decoded stuber.Output
	require.NoError(t, json.Unmarshal(jsonData, &decoded))
	require.Equal(t, output.Delay, decoded.Delay)
	require.Equal(t, output.DelayType, decoded.DelayType)
	require.Equal(t, output.DelayStep, decoded.DelayStep)
}

func TestOutputRegressiveDelayYAMLDeserialization(t *testing.T) {
	t.Parallel()

	data := []byte(`
service: example.Service
method: Get
input:
  contains:
    key: example
output:
  delay: 3s
  delay_type: regressive
  delay_step: 500ms
  data:
    result: success
`)

	var stub stuber.Stub

	jsonData, err := yaml.YAMLToJSON(data)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(jsonData, &stub))
	require.Equal(t, types.Duration(3*time.Second), stub.Output.Delay)
	require.Equal(t, stuber.DelayTypeRegressive, stub.Output.DelayType)
	require.Equal(t, types.Duration(500*time.Millisecond), stub.Output.DelayStep)
}
