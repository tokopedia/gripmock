package stub

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStub(t *testing.T) {
	type test struct {
		name    string
		mock    func() *http.Request
		handler http.HandlerFunc
		expect  string
	}

	cases := []test{
		{
			name: "add stub simple",
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
				return httptest.NewRequest("GET", "/", nil)
			},
			handler: listStub,
			expect:  "{\"Testing\":{\"TestMethod\":[{\"Input\":{\"equals\":{\"Hola\":\"Mundo\"},\"contains\":null,\"matches\":null},\"Output\":{\"data\":{\"Hello\":\"World\"},\"error\":\"\"}}]}}\n",
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
		}, {
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
		}, {
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
		}, {
			name: "error find stub contains",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing",
						"method":"TestMethod",
						"data":{
							"field1":"hello field1",
							"field2":"hello field2",
							"field3":"hello field4"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "Can't find stub \n\nService: Testing \n\nMethod: TestMethod \n\nInput\n\n{\n\tfield1: hello field1\n\tfield2: hello field2\n\tfield3: hello field4\n}\n\nClosest Match \n\ncontains:{\n\tfield1: hello field1\n\tfield3: hello field3\n}",
		}, {
			name: "error find stub equals",
			mock: func() *http.Request {
				payload := `{"service":"Testing","method":"TestMethod","data":{"Hola":"Dunia"}}`
				return httptest.NewRequest("POST", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "Can't find stub \n\nService: Testing \n\nMethod: TestMethod \n\nInput\n\n{\n\tHola: Dunia\n}\n\nClosest Match \n\nequals:{\n\tHola: Mundo\n}",
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			wrt := httptest.NewRecorder()
			req := v.mock()
			v.handler(wrt, req)
			res, err := ioutil.ReadAll(wrt.Result().Body)

			assert.NoError(t, err)
			assert.Equal(t, v.expect, string(res))
		})
	}
}
