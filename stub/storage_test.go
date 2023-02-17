package stub

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

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
