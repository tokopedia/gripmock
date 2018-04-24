package stub

import (
	"bytes"
	"io/ioutil"
	"log"
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
			expect: `{"success":"Success add stub"}
`,
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
			expect: `{"success":"Success add stub"}
`,
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
			name: "find stub contains error",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing",
						"method":"TestMethod",
						"data":{
							"field1":"helllo field1",
							"field2":"hello field2",
							"field3":"hello field31"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect: `{"error":"Can't find stub for Service:Testing Method:TestMethod Input:{\"field1\":\"helllo field1\",\"field2\":\"hello field2\",\"field3\":\"hello field31\"}"}
`,
		}, {
			name: "add stub matches field equals",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing2",
						"method":"TestMethod",
						"input":{
							"matches":{
								"field1":{
									"equals":"equals field1"
								}
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
			expect:  "{\"success\":\"Success add stub\"}\n",
		}, {
			name: "find stub matches field equals",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing2",
						"method":"TestMethod",
						"data":{
							"field1":"equals field1",
							"field2":"field2",
							"field3":"field31"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"reply\":\"OK\"},\"error\":\"\"}\n",
		}, {
			name: "find stub matches field equals error",
			mock: func() *http.Request {
				payload := `{
						"service":"Testing2",
						"method":"TestMethod",
						"data":{
							"field1":"equals field0",
							"field2":"field2",
							"field3":"field31"
						}
					}`
				return httptest.NewRequest("GET", "/find", bytes.NewReader([]byte(payload)))
			},
			handler: handleFindStub,
			expect:  "{\"data\":{\"reply\":\"OK\"},\"error\":\"\"}\n",
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			wrt := httptest.NewRecorder()
			req := v.mock()
			v.handler(wrt, req)
			res, err := ioutil.ReadAll(wrt.Result().Body)

			log.Println(err)
			assert.NoError(t, err)
			assert.Equal(t, v.expect, string(res))
		})
	}
}
