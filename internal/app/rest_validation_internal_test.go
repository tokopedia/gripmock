package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bavix/gripmock/v3/internal/infra/stuber"
)

// RestValidationTestSuite provides test suite for REST API validation.
type RestValidationTestSuite struct {
	suite.Suite

	server     *RestServer
	budgerigar *stuber.Budgerigar
}

// SetupSuite initializes the test suite.
func (s *RestValidationTestSuite) SetupSuite() {
	s.budgerigar = stuber.NewBudgerigar()
	extender := &mockExtender{}
	server, err := NewRestServer(s.T().Context(), s.budgerigar, extender, nil, nil, nil, nil)
	s.Require().NoError(err)
	s.server = server
}

// SetupTest cleans up before each test.
func (s *RestValidationTestSuite) SetupTest() {
	s.budgerigar.Clear()
}

// TestAddStubValidationErrors tests validation error cases for AddStub.
//
//nolint:funlen
func (s *RestValidationTestSuite) TestAddStubValidationErrors() {
	tests := []struct {
		name           string
		jsonData       string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing service field",
			jsonData: `[{
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "service name is missing",
		},
		{
			name: "missing method field",
			jsonData: `[{
				"service": "TestService",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "method name is missing",
		},
		{
			name: "both input and inputs provided (invalid configuration)",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"inputs": [{"contains": {"key": "value"}}],
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "both output.data and output.stream provided (invalid configuration)",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}, "stream": [{"result": "stream"}]}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'data' or 'stream', but not both",
		},
		{
			name: "unary stub without input",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "unary stub without output",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'data' or 'stream', but not both",
		},
		{
			name: "client streaming stub without inputs",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "server streaming stub without input",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"output": {"stream": [{"result": "response1"}]}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "empty service name",
			jsonData: `[{
				"service": "",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "service name is missing",
		},
		{
			name: "empty method name",
			jsonData: `[{
				"service": "TestService",
				"method": "",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "method name is missing",
		},
		{
			name: "input with all empty matchers",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": null, "equals": null, "matches": null},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "input with empty anyOf element",
			jsonData: `[{ 
				"service": "TestService",
				"method": "TestMethod",
				"input": {"anyOf": [{}]},
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'input' or 'inputs', but not both",
		},
		{
			name: "output with empty data and error",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": null, "error": ""}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must have either 'data' or 'stream', but not both",
		},
		{
			name: "invalid effect action",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"effects": [{"action": "unknown"}],
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid effects configuration",
		},
		{
			name: "delete effect without id",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"effects": [{"action": "delete"}],
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid effects configuration",
		},
		{
			name: "upsert effect without stub",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"effects": [{"action": "upsert"}],
				"output": {"data": {"result": "success"}}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid effects configuration",
		},
		{
			name: "regressive delay without step",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {
					"delay": "1s",
					"delay_type": "regressive",
					"data": {"result": "success"}
				}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid delay configuration",
		},
		{
			name: "delay step with default type",
			jsonData: `[{
				"service": "TestService",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {
					"delay": "1s",
					"delay_type": "default",
					"delay_step": "100ms",
					"data": {"result": "success"}
				}
			}]`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid delay configuration",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost, "/api/stubs", bytes.NewBufferString(tt.jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			s.server.AddStub(w, req)

			s.Equal(tt.expectedStatus, w.Code)
			s.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

// TestAddStubValidConfigurations tests valid stub configurations.
//
//nolint:funlen
func (s *RestValidationTestSuite) TestAddStubValidConfigurations() {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name: "valid unary stub with contains matcher",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid unary stub with equals matcher",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"equals": {"key": "exact_value"}},
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid unary stub with matches matcher",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"matches": {"key": "pattern.*"}},
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid unary stub with anyOf matcher",
			jsonData: `[{ 
				"service": "test.Service",
				"method": "TestMethod",
				"input": {
					"equals": {"role": "vip"},
					"anyOf": [
						{"equals": {"name": "Alice"}},
						{"matches": {"name": "^admin_"}}
					]
				},
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid unary stub with upsert effect",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"effects": [{
					"action": "upsert",
					"stub": {
						"service": "test.Service",
						"method": "GeneratedMethod",
						"input": {"equals": {"id": "1"}},
						"output": {"data": {"generated": true}}
					}
				}],
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid unary stub with error output",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"error": "Something went wrong"}
			}]`,
		},
		{
			name: "valid unary stub with code output",
			jsonData: `[{ 
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"code": 14, "error": "Service unavailable"}
			}]`,
		},
		{
			name: "valid unary stub with grpc details",
			jsonData: `[{ 
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {
					"code": 3,
					"error": "Validation failed",
					"details": [
						{
							"type": "type.googleapis.com/google.rpc.ErrorInfo",
							"reason": "API_DISABLED",
							"domain": "example.local"
						}
					]
				}
			}]`,
		},
		{
			name: "valid client streaming stub",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestClientStream",
				"inputs": [
					{"contains": {"key": "value1"}},
					{"contains": {"key": "value2"}}
				],
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid server streaming stub",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestServerStream",
				"input": {"contains": {"key": "value"}},
				"output": {"stream": [
					{"result": "response1"},
					{"result": "response2"}
				]}
			}]`,
		},
		{
			name: "valid bidirectional streaming stub",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestBidiStream",
				"inputs": [
					{"contains": {"key": "value1"}},
					{"contains": {"key": "value2"}}
				],
				"output": {"stream": [
					{"result": "response1"},
					{"result": "response2"}
				]}
			}]`,
		},
		{
			name: "valid stub with headers",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"headers": {"authorization": "Bearer token"},
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
		},
		{
			name: "valid stub with priority",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}},
				"priority": 10
			}]`,
		},
		{
			name: "valid stub with regressive delay",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {
					"delay": "1s",
					"delay_type": "regressive",
					"delay_step": "100ms",
					"data": {"result": "success"}
				}
			}]`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost, "/api/stubs", bytes.NewBufferString(tt.jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			s.server.AddStub(w, req)

			s.Equal(http.StatusOK, w.Code)
			s.NotEmpty(w.Body.String())
		})
	}
}

// TestAddStubInvalidJSON tests invalid JSON handling.
func (s *RestValidationTestSuite) TestAddStubInvalidJSON() {
	tests := []struct {
		name           string
		jsonData       string
		expectedStatus int
		description    string
	}{
		{
			name:           "invalid JSON syntax",
			jsonData:       `[{"service": "test", "method": "test", invalid}]`,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should reject malformed JSON",
		},
		{
			name:           "empty JSON",
			jsonData:       ``,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should handle empty body",
		},
		{
			name:           "not an array",
			jsonData:       `{"service": "test", "method": "test"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Should require array format",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost, "/api/stubs", bytes.NewBufferString(tt.jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			s.server.AddStub(w, req)

			s.Equal(tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// TestAddStubContentTypeValidation tests content type validation.
//
//nolint:funlen
func (s *RestValidationTestSuite) TestAddStubContentTypeValidation() {
	validStubData := `[{
		"service": "test.Service",
		"method": "TestMethod",
		"input": {"contains": {"key": "value"}},
		"output": {"data": {"result": "success"}}
	}]`

	tests := []struct {
		name        string
		contentType string
		expectPass  bool
		description string
	}{
		{
			name:        "valid content type",
			contentType: "application/json",
			expectPass:  true,
			description: "Should accept application/json",
		},
		{
			name:        "valid content type with charset",
			contentType: "application/json; charset=utf-8",
			expectPass:  true,
			description: "Should accept application/json with charset",
		},
		{
			name:        "missing content type",
			contentType: "",
			expectPass:  true,
			description: "Should accept missing content type if JSON is valid",
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			expectPass:  true,
			description: "Should accept wrong content type if JSON is valid",
		},
		{
			name:        "xml content type",
			contentType: "application/xml",
			expectPass:  true,
			description: "Should accept XML content type if JSON is valid",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost, "/api/stubs", bytes.NewBufferString(validStubData))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			s.server.AddStub(w, req)

			if tt.expectPass {
				s.Equal(http.StatusOK, w.Code, tt.description)
			} else {
				s.Equal(http.StatusBadRequest, w.Code, tt.description)
			}
		})
	}
}

// TestAddStubSpecialCases tests edge cases and special configurations.
func (s *RestValidationTestSuite) TestAddStubSpecialCases() {
	tests := []struct {
		name        string
		jsonData    string
		description string
	}{
		{
			name: "server streaming stub with data output (valid but may be treated as unary)",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"input": {"contains": {"key": "value"}},
				"output": {"data": {"result": "success"}}
			}]`,
			description: "Server streaming can have data output",
		},
		{
			name: "bidirectional streaming stub with data output (valid but may be treated as client streaming)",
			jsonData: `[{
				"service": "test.Service",
				"method": "TestMethod",
				"inputs": [{"contains": {"key": "value"}}],
				"output": {"data": {"result": "success"}}
			}]`,
			description: "Bidirectional streaming can have data output",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequestWithContext(s.T().Context(), http.MethodPost, "/api/stubs", bytes.NewBufferString(tt.jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			s.server.AddStub(w, req)

			s.Equal(http.StatusOK, w.Code, tt.description)
			s.NotEmpty(w.Body.String())
		})
	}
}

// TestRestValidationTestSuite runs the REST validation test suite.
func TestRestValidationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RestValidationTestSuite))
}
