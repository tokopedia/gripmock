package stub

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_findStub(t *testing.T) {
	tests := []struct {
		name          string
		setup         *Stub
		input         *findStubPayload
		wantOutput    *Output
		wantErr       bool
		wantErrString string
	}{
		{
			name: "service not found",
			input: &findStubPayload{
				Service: "nonexistent",
				Method:  "test",
			},
			wantErr:       true,
			wantErrString: "can't find stub for Service: nonexistent",
		},
		{
			name: "method not found",
			setup: &Stub{
				Service: "test",
				Method:  "method1",
			},
			input: &findStubPayload{
				Service: "test",
				Method:  "method2",
			},
			wantErr:       true,
			wantErrString: "can't find stub for Service:test and Method:method2",
		},
		{
			name: "exact match - simple",
			setup: &Stub{
				Service: "user",
				Method:  "GetUser",
				Input: Input{
					Equals: map[string]interface{}{
						"id": 1,
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"name": "John",
					},
				},
			},
			input: &findStubPayload{
				Service: "user",
				Method:  "GetUser",
				Data: map[string]interface{}{
					"id": 1,
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"name": "John",
				},
			},
		},
		{
			name: "exact match - nested structure",
			setup: &Stub{
				Service: "user",
				Method:  "GetUser",
				Input: Input{
					Equals: map[string]interface{}{
						"user": map[string]interface{}{
							"id":   1,
							"type": "admin",
						},
						"options": []interface{}{
							"full",
							"details",
						},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"result": "success",
					},
				},
			},
			input: &findStubPayload{
				Service: "user",
				Method:  "GetUser",
				Data: map[string]interface{}{
					"user": map[string]interface{}{
						"id":   1,
						"type": "admin",
					},
					"options": []interface{}{
						"full",
						"details",
					},
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"result": "success",
				},
			},
		},
		{
			name: "equals unordered - array elements",
			setup: &Stub{
				Service: "test",
				Method:  "Test",
				Input: Input{
					EqualsUnordered: map[string]interface{}{
						"tags": []interface{}{"a", "b", "c"},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"result": "found",
					},
				},
			},
			input: &findStubPayload{
				Service: "test",
				Method:  "Test",
				Data: map[string]interface{}{
					"tags": []interface{}{"c", "a", "b"},
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"result": "found",
				},
			},
		},
		{
			name: "contains match",
			setup: &Stub{
				Service: "product",
				Method:  "Search",
				Input: Input{
					Contains: map[string]interface{}{
						"category": "electronics",
						"price": map[string]interface{}{
							"min": 100,
						},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"products": []interface{}{
							map[string]interface{}{"id": 1},
						},
					},
				},
			},
			input: &findStubPayload{
				Service: "product",
				Method:  "Search",
				Data: map[string]interface{}{
					"category": "electronics",
					"price": map[string]interface{}{
						"min": 100,
						"max": 200,
					},
					"brand": "apple",
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"products": []interface{}{
						map[string]interface{}{"id": 1},
					},
				},
			},
		},
		{
			name: "regex match",
			setup: &Stub{
				Service: "validation",
				Method:  "ValidateEmail",
				Input: Input{
					Matches: map[string]interface{}{
						"email": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"valid": true,
					},
				},
			},
			input: &findStubPayload{
				Service: "validation",
				Method:  "ValidateEmail",
				Data: map[string]interface{}{
					"email": "test@example.com",
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"valid": true,
				},
			},
		},
		{
			name: "headers - exact match",
			setup: &Stub{
				Service: "auth",
				Method:  "Verify",
				Input: Input{
					Equals: map[string]interface{}{
						"token": "123",
					},
					Headers: &InputHeaders{
						Equals: map[string]string{
							"Authorization": "Bearer token123",
							"X-Request-ID":  "abc123",
						},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"valid": true,
					},
					Headers: map[string]string{
						"X-Response-ID": "xyz789",
					},
				},
			},
			input: &findStubPayload{
				Service: "auth",
				Method:  "Verify",
				Data: map[string]interface{}{
					"token": "123",
				},
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Request-ID":  "abc123",
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"valid": true,
				},
				Headers: map[string]string{
					"X-Response-ID": "xyz789",
				},
			},
		},
		{
			name: "headers - contains match",
			setup: &Stub{
				Service: "auth",
				Method:  "Verify",
				Input: Input{
					Equals: map[string]interface{}{
						"token": "123",
					},
					Headers: &InputHeaders{
						Contains: map[string]string{
							"Authorization": "Bearer",
						},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"valid": true,
					},
				},
			},
			input: &findStubPayload{
				Service: "auth",
				Method:  "Verify",
				Data: map[string]interface{}{
					"token": "123",
				},
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Extra":       "value",
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"valid": true,
				},
			},
		},
		{
			name: "headers - regex match",
			setup: &Stub{
				Service: "auth",
				Method:  "Verify",
				Input: Input{
					Equals: map[string]interface{}{
						"token": "123",
					},
					Headers: &InputHeaders{
						Matches: map[string]string{
							"X-Request-ID": "^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$",
						},
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"valid": true,
					},
				},
			},
			input: &findStubPayload{
				Service: "auth",
				Method:  "Verify",
				Data: map[string]interface{}{
					"token": "123",
				},
				Headers: map[string]string{
					"X-Request-ID": "550e8400-e29b-41d4-a716-446655440000",
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"valid": true,
				},
			},
		},
		{
			name: "multiple stubs - most specific match",
			setup: &Stub{
				Service: "user",
				Method:  "GetUser",
				Input: Input{
					Contains: map[string]interface{}{
						"id": 1,
					},
				},
				Output: Output{
					Data: map[string]interface{}{
						"name": "John Generic",
					},
				},
			},
			input: &findStubPayload{
				Service: "user",
				Method:  "GetUser",
				Data: map[string]interface{}{
					"id":      1,
					"details": true,
				},
			},
			wantOutput: &Output{
				Data: map[string]interface{}{
					"name": "John Generic",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear storage before each test
			clearStorage()

			// Setup stub if provided
			if tt.setup != nil {
				err := storeStub(tt.setup)
				require.NoError(t, err)
			}

			// Execute test
			got, err := findStub(tt.input)

			// Verify error cases
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrString != "" {
					require.Contains(t, err.Error(), tt.wantErrString)
				}
				return
			}

			// Verify success cases
			require.NoError(t, err)
			require.NotNil(t, got)

			// Check output data
			if tt.wantOutput.Data != nil {
				require.True(t, reflect.DeepEqual(tt.wantOutput.Data, got.Data),
					"Expected output data %v, got %v", tt.wantOutput.Data, got.Data)
			}

			// Check output headers
			if tt.wantOutput.Headers != nil {
				require.True(t, reflect.DeepEqual(tt.wantOutput.Headers, got.Headers),
					"Expected output headers %v, got %v", tt.wantOutput.Headers, got.Headers)
			}
		})
	}
}

func Test_readStubFromFile(t *testing.T) {
	tests := []struct {
		name        string
		mock        func(service, method string, data []storage) (path string)
		service     string
		method      string
		data        []storage
		expectCount int
	}{
		{
			name: "single file, single stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF.Close()

				var stubs []Stub
				for _, d := range data {
					stubs = append(stubs, Stub{
						Service: service,
						Method:  method,
						Input:   d.Input,
						Output:  d.Output,
					})
				}
				byt, err := json.Marshal(stubs)
				require.NoError(t, err)
				_, err = tempF.Write(byt)
				require.NoError(t, err)

				return dir
			},
			service: "user",
			method:  "getname",
			data: []storage{
				{
					Input:  Input{Equals: map[string]interface{}{"id": float64(1)}},
					Output: Output{Data: map[string]interface{}{"name": "user1"}},
				},
			},
			expectCount: 1,
		},
		{
			name: "single file, multiple stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF.Close()

				var stubs []Stub
				for _, d := range data {
					stubs = append(stubs, Stub{
						Service: service,
						Method:  method,
						Input:   d.Input,
						Output:  d.Output,
					})
				}
				byt, err := json.Marshal(stubs)
				require.NoError(t, err)
				_, err = tempF.Write(byt)
				require.NoError(t, err)

				return dir
			},
			service: "user",
			method:  "getname",
			data: []storage{
				{
					Input:  Input{Equals: map[string]interface{}{"id": float64(1)}},
					Output: Output{Data: map[string]interface{}{"name": "user1"}},
				},
				{
					Input:  Input{Equals: map[string]interface{}{"id": float64(2)}},
					Output: Output{Data: map[string]interface{}{"name": "user2"}},
				},
			},
			expectCount: 2,
		},
		{
			name: "multiple file, single stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)

				for _, d := range data {
					tempF, err := ioutil.TempFile(dir, "stub*.json")
					require.NoError(t, err)
					defer tempF.Close()

					stub := Stub{
						Service: service,
						Method:  method,
						Input:   d.Input,
						Output:  d.Output,
					}
					byt, err := json.Marshal(stub)
					require.NoError(t, err)
					_, err = tempF.Write(byt)
					require.NoError(t, err)
				}

				return dir
			},
			service: "user",
			method:  "getname",
			data: []storage{
				{
					Input:  Input{Equals: map[string]interface{}{"id": float64(1)}},
					Output: Output{Data: map[string]interface{}{"name": "user1"}},
				},
				{
					Input:  Input{Equals: map[string]interface{}{"id": float64(2)}},
					Output: Output{Data: map[string]interface{}{"name": "user2"}},
				},
			},
			expectCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := stubMapping{}
			count := sm.readStubFromFile(tt.mock(tt.service, tt.method, tt.data))
			require.Equal(t, tt.expectCount, count)
			require.ElementsMatch(t, tt.data, sm[tt.service][tt.method])
		})
	}
}
