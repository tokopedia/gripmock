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
		name                  string
		service               string
		method                string
		stubInput             Input
		stubOutput            Output
		input                 map[string]interface{}
		inputHeaders          map[string][]string
		checkHeaders          bool
		expectedOutput        map[string]interface{}
		expectedOutputHeaders map[string][]string
	}{
		{
			name:           "input equals",
			service:        "user",
			method:         "getName",
			stubInput:      Input{Equals: map[string]interface{}{"id": float64(1)}},
			stubOutput:     Output{Data: map[string]interface{}{"name": "user1"}},
			input:          map[string]interface{}{"id": 1},
			checkHeaders:   false,
			expectedOutput: map[string]interface{}{"name": "user1"},
		},
		/*
					{
						name: "input contains",
			                        service: "user",
			                        method: "getName",
			                        stubInput: Input{Equals: map[string]interface{}{"id": float64(1)}},
						stubOutput: Output{Data: map[string]interface{}{"name": "user1"}},
			                        input: map[string]interface{} {
			                                "id": 1,
			                        },
			                        checkHeaders: false,
			                        expectedOutput: map[string]interface{} {
			                                "name": "user1",
			                        },
					},
					{
						name: "input matches",
			                        service: "user",
			                        method: "getName",
			                        stubInput: Input{Equals: map[string]interface{}{"id": float64(1)}},
						stubOutput: Output{Data: map[string]interface{}{"name": "user1"}},
			                        input: map[string]interface{} {
			                                "id": 1,
			                        },
			                        checkHeaders: false,
			                        expectedOutput: map[string]interface{} {
			                                "name": "user1",
			                        },
					},
					{
						name: "input equals and input headers equals",
			                        service: "user",
			                        method: "getName",
			                        stubInput: Input{
			                                Equals: map[string]interface{}{"id": float64(1)}},
			                                CheckHeaders: true,
			                                EqulsHeaders: map[string][]string{
			                                        "header-1": []string{"value-1", "value-2"},
			                                        "header-2": []string{"value-3", "value-4"},
			                                },
			                        },
						stubOutput: Output{
			                                Data: map[string]interface{}{"name": "user1"},
			                                Headers: map[string][]string{
			                                        "return-header": []string{"value-1", value-2"},
			                                },
			                        },
			                        input: map[string]interface{} {
			                                "id": 1,
			                        },
			                        inputHeaders: map[string][]string{
			                                "header-1": []string{"value-1", "value-2"},
			                                "header-2": []string{"value-3", "value-4"},
			                        },
			                        checkHeaders: true,
			                        expectedOutput: map[string]interface{} {
			                                "name": "user1",
			                        },
			                        expectedOutputHeaders: map[string][]string{
			                                "return-header": []string{"value-1", value-2"},
			                        },
					},
					{
						name: "input equals and input headers conatin",
			                        service: "user",
			                        method: "getName",
			                        stubInput: Input{
			                                Equals: map[string]interface{}{"id": float64(1)}},
			                                CheckHeaders: true,
			                                ContainsHeaders: map[string][]string{
			                                        "header-1": []string{"value-1"},
			                                        "header-2": []string{"value-4"},
			                                },
			                        },
						stubOutput: Output{
			                                Data: map[string]interface{}{"name": "user1"},
			                                Headers: map[string][]string{
			                                        "return-header": []string{"value-1", value-2"},
			                                },
			                        },
			                        input: map[string]interface{} {
			                                "id": 1,
			                        },
			                        inputHeaders: map[string][]string{
			                                "header-1": []string{"value-1"},
			                                "header-2": []string{"value-4"},
			                        },
			                        checkHeaders: true,
			                        expectedOutput: map[string]interface{} {
			                                "name": "user1",
			                        },
			                        expectedOutputHeaders: map[string][]string{
			                                "return-header": []string{"value-1", value-2"},
			                        },
					},
					{
						name: "input equals and input headers match",
			                        service: "user",
			                        method: "getName",
			                        stubInput: Input{
			                                Equals: map[string]interface{}{"id": float64(1)}},
			                                CheckHeaders: true,
			                                MatchesHeaders: map[string][]string{
			                                        "header-1": []string{"value-.*", "value-.*"},
			                                        "header-.*": []string{".*"},
			                                },
			                        },
						stubOutput: Output{
			                                Data: map[string]interface{}{"name": "user1"},
			                                Headers: map[string][]string{
			                                        "return-header": []string{"value-1", value-2"},
			                                },
			                        },
			                        input: map[string]interface{} {
			                                "id": 1,
			                        },
			                        inputHeaders: map[string][]string{
			                                "header-1": []string{"value-1", "value-2"},
			                                "header-2": []string{"value-3", "value-4"},
			                                "header-3": []string{"value-5", "value-6"},
			                        },
			                        checkHeaders: true,
			                        expectedOutput: map[string]interface{} {
			                                "name": "user1",
			                        },
			                        expectedOutputHeaders: map[string][]string{
			                                "return-header": []string{"value-1", value-2"},
			                        },
					},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := stubMapping{}
			err := sm.storeStub(&Stub{
				Service: tt.service,
				Method:  tt.method,
				Input:   tt.stubInput,
				Output:  tt.stubOutput,
			})
			require.NoError(t, err)

			output, err := findStub(&findStubPayload{
				Service: tt.service,
				Method:  tt.method,
				Data:    tt.input,
				Headers: tt.inputHeaders,
			})
			require.NoError(t, err)

			require.True(t, reflect.DeepEqual(tt.expectedOutput, output.Data))
			require.True(t, reflect.DeepEqual(tt.expectedOutputHeaders, output.Headers))
		})
	}
}

func Test_readStubFromFile(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(service, method string, data []storage) (path string)
		service string
		method  string
		data    []storage
	}{
		{
			name: "single file, single stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "")
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
		},
		{
			name: "single file, multiple stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)
				tempF, err := ioutil.TempFile(dir, "")
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
		},
		{
			name: "multiple file, single stub",
			mock: func(service, method string, data []storage) (path string) {
				dir, err := ioutil.TempDir("", "")
				require.NoError(t, err)

				for _, d := range data {
					tempF, err := ioutil.TempFile(dir, "")
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := stubMapping{}
			sm.readStubFromFile(tt.mock(tt.service, tt.method, tt.data))
			require.ElementsMatch(t, tt.data, sm[tt.service][tt.method])
		})
	}
}
