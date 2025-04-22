package stub

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStub(t *testing.T) {
	type test struct {
		name    string
		mock    func() *http.Request
		handler http.HandlerFunc
		expect  string
		verify  func(t *testing.T, w *httptest.ResponseRecorder)
		cleanup func(t *testing.T)
	}

	cases := []test{
		{
			name: "add simple stub",
			mock: func() *http.Request {
				payload := `{
						"service": "Testing",
						"method":"TestMethod",
						"input":{
							"equals":{
								"Hola":"Mundo"
							}
						},
						"output":{
							"data":{
								"Hello":"World"
							}
						}
					}`
				read := bytes.NewReader([]byte(payload))
				return httptest.NewRequest("POST", "/add", read)
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "list stub",
			mock: func() *http.Request {
				clearStorage()
				// Add the test stub
				stub := &Stub{
					Service: "Testing",
					Method:  "TestMethod",
					Input: Input{
						Equals: map[string]interface{}{
							"Hola": "Mundo",
						},
					},
					Output: Output{
						Data: map[string]interface{}{
							"Hello": "World",
						},
					},
				}
				err := storeStub(stub)
				if err != nil {
					panic(err)
				}
				return httptest.NewRequest("GET", "/", nil)
			},
			handler: listStub,
			expect:  "{\"Testing\":{\"TestMethod\":[{\"Input\":{\"equals\":{\"Hola\":\"Mundo\"},\"equals_unordered\":null,\"contains\":null,\"matches\":null},\"Output\":{\"data\":{\"Hello\":\"World\"},\"error\":\"\"}}]}}\n",
		},
		{
			name: "find stub equals",
			mock: func() *http.Request {
				payload := `{"service":"Testing","method":"TestMethod","data":{"Hola":"Mundo"}}`
				return httptest.NewRequest("POST", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"Hello\":\"World\"},\"error\":\"\"}\n",
		},
		{
			name: "add nested stub equals",
			mock: func() *http.Request {
				payload := `{
						"service": "NestedTesting",
						"method":"TestMethod",
						"input":{
							"equals":{
										"name": "Afra Gokce",
										"age": 1,
										"girl": true,
										"null": null,
										"greetings": {
											"hola": "mundo",
											"merhaba": "dunya"
										},
										"cities": ["Istanbul", "Jakarta"]
							}
						},
						"output":{
							"data":{
								"Hello":"World"
							}
						}
					}`
				read := bytes.NewReader([]byte(payload))
				return httptest.NewRequest("POST", "/add", read)
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "find nested stub equals",
			mock: func() *http.Request {
				payload := `{"service":"NestedTesting","method":"TestMethod","data":{"name":"Afra Gokce","age":1,"girl":true,"null":null,"greetings":{"hola":"mundo","merhaba":"dunya"},"cities":["Istanbul","Jakarta"]}}`
				return httptest.NewRequest("POST", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"Hello\":\"World\"},\"error\":\"\"}\n",
		},
		{
			name: "get recorded requests",
			mock: func() *http.Request {
				return httptest.NewRequest("GET", "/requests", nil)
			},
			handler: listRequests,
			expect:  "[{\"record\":{\"service\":\"Testing\",\"method\":\"TestMethod\",\"data\":{\"Hola\":\"Mundo\"}},\"count\":1},{\"record\":{\"service\":\"NestedTesting\",\"method\":\"TestMethod\",\"data\":{\"age\":1,\"cities\":[\"Istanbul\",\"Jakarta\"],\"girl\":true,\"greetings\":{\"hola\":\"mundo\",\"merhaba\":\"dunya\"},\"name\":\"Afra Gokce\",\"null\":null}},\"count\":1}]\n",
		},
		{
			name: "add stub equals_unordered",
			mock: func() *http.Request {
				payload := `{
								"service": "TestingUnordered",
								"method":"TestMethod",
								"input": {
									"equals_unordered": {
										"ids": [1,2]
									}
								},
								"output":{
									"data":{
										"hello":"world"
									}
								}
							}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "find stub equals_unordered",
			mock: func() *http.Request {
				payload := `{
						"service":"TestingUnordered",
						"method":"TestMethod",
						"data":{
							"ids":[1,2]
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"hello\":\"world\"},\"error\":\"\"}\n",
		},
		{
			name: "find stub equals_unordered reversed",
			mock: func() *http.Request {
				payload := `{
						"service":"TestingUnordered",
						"method":"TestMethod",
						"data":{
							"ids":[2,1]
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"hello\":\"world\"},\"error\":\"\"}\n",
		},
		{
			name: "add stub contains",
			mock: func() *http.Request {
				payload := `{
								"service": "Testing",
								"method":"TestMethod",
								"input":{
									"contains":{
										"field1":"hello field1",
										"field3":"hello field3"
									}
								},
								"output":{
									"data":{
										"hello":"world"
									}
								}
							}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "find stub contains",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing",
						"method":"TestMethod",
						"data":{
							"field1":"hello field1",
							"field2":"hello field2",
							"field3":"hello field3"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"hello\":\"world\"},\"error\":\"\"}\n",
		},
		{
			name: "add nested stub contains",
			mock: func() *http.Request {
				payload := `{
								"service": "NestedTesting",
								"method":"TestMethod",
								"input":{
									"contains":{
												"key": "value",
												"greetings": {
													"hola": "mundo",
													"merhaba": "dunya"
												},
												"cities": ["Istanbul", "Jakarta"]
									}
								},
								"output":{
									"data":{
										"hello":"world"
									}
								}
							}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "add error stub with result code contains",
			mock: func() *http.Request {
				payload := `{
								"service": "ErrorStabWithCode",
								"method":"TestMethod",
								"input":{
									"contains":{
												"key": "value",
												"greetings": {
													"hola": "mundo",
													"merhaba": "dunya"
												},
												"cities": ["Istanbul", "Jakarta"]
									}
								},
								"output":{
									"error":"error msg",
                                    "code": 3
								}
							}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "find error stub with result code contains",
			mock: func() *http.Request {
				payload := `{
						"service": "ErrorStabWithCode",
						"method":"TestMethod",
						"data":{
								"key": "value",
								"anotherKey": "anotherValue",
								"greetings": {
									"hola": "mundo",
									"merhaba": "dunya",
									"hello": "world"
								},
								"cities": ["Istanbul", "Jakarta", "Winterfell"]
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":null,\"error\":\"error msg\",\"code\":3}\n",
		},

		{
			name: "add error stub without result code contains",
			mock: func() *http.Request {
				payload := `{
								"service": "ErrorStab",
								"method":"TestMethod",
								"input":{
									"contains":{
												"key": "value",
												"greetings": {
													"hola": "mundo",
													"merhaba": "dunya"
												},
												"cities": ["Istanbul", "Jakarta"]
									}
								},
								"output":{
									"error":"error msg"
								}
							}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  `Success add stub`,
		},
		{
			name: "find error stub without result code contains",
			mock: func() *http.Request {
				payload := `{
						"service": "ErrorStab",
						"method":"TestMethod",
						"data":{
								"key": "value",
								"anotherKey": "anotherValue",
								"greetings": {
									"hola": "mundo",
									"merhaba": "dunya",
									"hello": "world"
								},
								"cities": ["Istanbul", "Jakarta", "Winterfell"]
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":null,\"error\":\"error msg\"}\n",
		},
		{
			name: "find nested stub contains",
			mock: func() *http.Request {
				payload := `{
						"service":"NestedTesting",
						"method":"TestMethod",
						"data":{
								"key": "value",
								"anotherKey": "anotherValue",
								"greetings": {
									"hola": "mundo",
									"merhaba": "dunya",
									"hello": "world"
								},
								"cities": ["Istanbul", "Jakarta", "Winterfell"]
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"hello\":\"world\"},\"error\":\"\"}\n",
		},
		{
			name: "add stub matches regex",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing2",
						"method":"TestMethod",
						"input":{
							"matches":{
								"field1":".*ello$"
							}
						},
						"output":{
							"data":{
								"reply":"OK"
							}
						}
					}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  "Success add stub",
		},
		{
			name: "find stub matches regex",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing2",
						"method":"TestMethod",
						"data":{
							"field1":"hello"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"reply\":\"OK\"},\"error\":\"\"}\n",
		},
		{
			name: "add nested stub matches regex",
			mock: func() *http.Request {
				payload := `{
						"service":"NestedTesting2",
						"method":"TestMethod",
						"input":{
							"matches":{
										"key": "[a-z]{3}ue",
										"greetings": {
											"hola": 1,
											"merhaba": true,
											"hello": "^he[l]{2,}o$"
										},
										"cities": ["Istanbul", "Jakarta", ".*"],
										"mixed": [5.5, false, ".*"]
							}
						},
						"output":{
							"data":{
								"reply":"OK"
							}
						}
					}`
				return httptest.NewRequest("POST", "/add", bytes.NewReader([]byte(payload)))
			},
			handler: addStub,
			expect:  "Success add stub",
		},
		{
			name: "find nested stub matches regex",
			mock: func() *http.Request {
				payload := `{
						"service":"NestedTesting2",
						"method":"TestMethod",
						"data":{
								"key": "value",
								"greetings": {
									"hola": 1,
									"merhaba": true,
									"hello": "helllllo"
								},
								"cities": ["Istanbul", "Jakarta", "Gotham"],
								"mixed": [5.5, false, "Gotham"]
							}
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"reply\":\"OK\"},\"error\":\"\"}\n",
		},
		{
			name: "error_find_stub_contains",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing",
						"method":"TestMethod",
						"data":{
							"field1":"hello field1"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Instead of checking the exact error message, verify it contains the expected fields
				res, err := ioutil.ReadAll(w.Result().Body)
				assert.NoError(t, err)
				errMsg := string(res)
				assert.Contains(t, errMsg, "Service: Testing")
				assert.Contains(t, errMsg, "Method: TestMethod")
				assert.Contains(t, errMsg, "field1: hello field1")
				assert.Contains(t, errMsg, "field3: hello field3")
			},
		},
		{
			name: "error find stub equals",
			mock: func() *http.Request {
				payload := `{"service":"Testing","method":"TestMethod","data":{"Hola":"Dunia"}}`
				return httptest.NewRequest("POST", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "Can't find stub \n\nService: Testing \n\nMethod: TestMethod \n\nInput\n\nData:\n{\n\tHola: Dunia\n}\n\nClosest Match \n\nequals:{\n\tHola: Mundo\n}",
		},
		{
			name: "reset stubs with path configured",
			mock: func() *http.Request {
				// Set up a temporary directory with a stub file
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF.Close()

				stub := Stub{
					Service: "TestService",
					Method:  "TestMethod",
					Input:   Input{Equals: map[string]interface{}{"field": "value"}},
					Output:  Output{Data: map[string]interface{}{"result": "success"}},
				}
				byt, err := json.Marshal(stub)
				require.NoError(t, err)
				_, err = tempF.Write(byt)
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 1 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify the stub was loaded
				stubs := allStub()
				assert.Contains(t, stubs, "TestService")
				assert.Contains(t, stubs["TestService"], "TestMethod")
				assert.Len(t, stubs["TestService"]["TestMethod"], 1)
				assert.Equal(t, map[string]interface{}{"field": "value"}, stubs["TestService"]["TestMethod"][0].Input.Equals)
				assert.Equal(t, map[string]interface{}{"result": "success"}, stubs["TestService"]["TestMethod"][0].Output.Data)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
		{
			name: "reset stubs with array of stubs",
			mock: func() *http.Request {
				// Set up a temporary directory with a stub file containing array of stubs
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF.Close()

				stubs := []Stub{
					{
						Service: "Service1",
						Method:  "Method1",
						Input:   Input{Equals: map[string]interface{}{"field1": "value1"}},
						Output:  Output{Data: map[string]interface{}{"result1": "success1"}},
					},
					{
						Service: "Service2",
						Method:  "Method2",
						Input:   Input{Equals: map[string]interface{}{"field2": "value2"}},
						Output:  Output{Data: map[string]interface{}{"result2": "success2"}},
					},
				}
				byt, err := json.Marshal(stubs)
				require.NoError(t, err)
				_, err = tempF.Write(byt)
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 2 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify both stubs were loaded
				stubs := allStub()
				assert.Contains(t, stubs, "Service1")
				assert.Contains(t, stubs, "Service2")
				assert.Contains(t, stubs["Service1"], "Method1")
				assert.Contains(t, stubs["Service2"], "Method2")
				assert.Len(t, stubs["Service1"]["Method1"], 1)
				assert.Len(t, stubs["Service2"]["Method2"], 1)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
		{
			name: "reset stubs with invalid JSON",
			mock: func() *http.Request {
				// Set up a temporary directory with an invalid JSON file
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF.Close()

				// Write invalid JSON
				_, err = tempF.WriteString(`{"invalid": json}`)
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 0 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify storage is empty since invalid JSON was skipped
				stubs := allStub()
				assert.Empty(t, stubs)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
		{
			name: "reset stubs with multiple files",
			mock: func() *http.Request {
				// Set up a temporary directory with multiple stub files
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)

				// First stub file
				stub1 := Stub{
					Service: "Service1",
					Method:  "Method1",
					Input:   Input{Equals: map[string]interface{}{"field1": "value1"}},
					Output:  Output{Data: map[string]interface{}{"result1": "success1"}},
				}
				byt1, err := json.Marshal(stub1)
				require.NoError(t, err)
				err = ioutil.WriteFile(dir+"/stub1.json", byt1, 0644)
				require.NoError(t, err)

				// Second stub file
				stub2 := Stub{
					Service: "Service2",
					Method:  "Method2",
					Input:   Input{Equals: map[string]interface{}{"field2": "value2"}},
					Output:  Output{Data: map[string]interface{}{"result2": "success2"}},
				}
				byt2, err := json.Marshal(stub2)
				require.NoError(t, err)
				err = ioutil.WriteFile(dir+"/stub2.json", byt2, 0644)
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 2 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify both stubs were loaded
				stubs := allStub()
				assert.Contains(t, stubs, "Service1")
				assert.Contains(t, stubs, "Service2")
				assert.Contains(t, stubs["Service1"], "Method1")
				assert.Contains(t, stubs["Service2"], "Method2")
				assert.Len(t, stubs["Service1"]["Method1"], 1)
				assert.Len(t, stubs["Service2"]["Method2"], 1)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
		{
			name: "reset stubs without path configured",
			mock: func() *http.Request {
				// Clear the stub path
				stubPath = ""
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "No stub path configured",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify storage is empty
				stubs := allStub()
				assert.Empty(t, stubs)
			},
		},
		{
			name: "reset stubs with non-json files",
			mock: func() *http.Request {
				// Set up a temporary directory with a non-json file
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)

				// Create a .txt file (should be ignored)
				tempF1, err := ioutil.TempFile(dir, "stub*.txt")
				require.NoError(t, err)
				defer tempF1.Close()

				// Create a valid JSON file
				tempF2, err := ioutil.TempFile(dir, "stub*.json")
				require.NoError(t, err)
				defer tempF2.Close()

				stub := Stub{
					Service: "TestService",
					Method:  "TestMethod",
					Input:   Input{Equals: map[string]interface{}{"field": "value"}},
					Output:  Output{Data: map[string]interface{}{"result": "success"}},
				}
				byt, err := json.Marshal(stub)
				require.NoError(t, err)
				_, err = tempF2.Write(byt)
				require.NoError(t, err)

				// Write some non-JSON content to the .txt file
				_, err = tempF1.WriteString("This is not a JSON file")
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 1 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify only the JSON file was loaded
				stubs := allStub()
				assert.Contains(t, stubs, "TestService")
				assert.Contains(t, stubs["TestService"], "TestMethod")
				assert.Len(t, stubs["TestService"]["TestMethod"], 1)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
		{
			name: "reset stubs with nested directories",
			mock: func() *http.Request {
				// Set up a temporary directory with nested directories
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)

				// Create a subdirectory
				subdir := filepath.Join(dir, "subdir")
				err = os.Mkdir(subdir, 0755)
				require.NoError(t, err)

				// Create a stub file in the root directory
				stub1 := Stub{
					Service: "Service1",
					Method:  "Method1",
					Input:   Input{Equals: map[string]interface{}{"field1": "value1"}},
					Output:  Output{Data: map[string]interface{}{"result1": "success1"}},
				}
				byt1, err := json.Marshal(stub1)
				require.NoError(t, err)
				err = ioutil.WriteFile(filepath.Join(dir, "stub1.json"), byt1, 0644)
				require.NoError(t, err)

				// Create a stub file in the subdirectory
				stub2 := Stub{
					Service: "Service2",
					Method:  "Method2",
					Input:   Input{Equals: map[string]interface{}{"field2": "value2"}},
					Output:  Output{Data: map[string]interface{}{"result2": "success2"}},
				}
				byt2, err := json.Marshal(stub2)
				require.NoError(t, err)
				err = ioutil.WriteFile(filepath.Join(subdir, "stub2.json"), byt2, 0644)
				require.NoError(t, err)

				// Store the path for the handler to use
				stubPath = dir
				return httptest.NewRequest("POST", "/reset", nil)
			},
			handler: handleResetStub,
			expect:  "Stubs reset from files. Loaded 2 stubs.",
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Verify both stubs were loaded
				stubs := allStub()
				assert.Contains(t, stubs, "Service1")
				assert.Contains(t, stubs, "Service2")
				assert.Contains(t, stubs["Service1"], "Method1")
				assert.Contains(t, stubs["Service2"], "Method2")
				assert.Len(t, stubs["Service1"]["Method1"], 1)
				assert.Len(t, stubs["Service2"]["Method2"], 1)
			},
			cleanup: func(t *testing.T) {
				os.RemoveAll(stubPath)
			},
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			wrt := httptest.NewRecorder()
			req := v.mock()
			v.handler(wrt, req)

			if v.verify != nil {
				v.verify(t, wrt)
			} else if v.expect != "" {
				res, err := ioutil.ReadAll(wrt.Result().Body)
				assert.NoError(t, err)
				assert.Equal(t, v.expect, string(res))
			}

			if v.cleanup != nil {
				v.cleanup(t)
			}
		})
	}
}
