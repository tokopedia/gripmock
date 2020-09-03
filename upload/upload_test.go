package upload

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/quintans/gripmock/servers"
	"github.com/quintans/gripmock/servers/mocks"
	"github.com/quintans/gripmock/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploader(t *testing.T) {
	type test struct {
		name         string
		rebooterMock func(t *testing.T) (*mocks.MockRebooter, *gomock.Controller)
		mock         func(t *testing.T) *http.Request
		handler      func(us uploadServer, w http.ResponseWriter, r *http.Request)
		expect       string
	}

	cases := []test{
		{
			name: "upload ok",
			rebooterMock: func(t *testing.T) (*mocks.MockRebooter, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				rebooter := mocks.NewMockRebooter(ctrl)
				rebooter.EXPECT().UploadDir().Return("test-upload-dir").Times(1)
				rebooter.EXPECT().Shutdown().Times(1)
				rebooter.EXPECT().Boot(gomock.Any()).Times(1)
				return rebooter, ctrl
			},
			mock: func(t *testing.T) *http.Request {
				payload, err := tool.ZipFolder("../example/upload/proto")
				require.NoError(t, err)
				return httptest.NewRequest("POST", "/upload", bytes.NewReader([]byte(payload)))
			},
			handler: func(us uploadServer, w http.ResponseWriter, r *http.Request) {
				// created when downloading the zip
				defer os.RemoveAll("test-upload-dir")

				us.handleUpload(w, r)
			},
			expect: "",
		},
		{
			name: "reset import dirs true",
			rebooterMock: func(t *testing.T) (*mocks.MockRebooter, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				rebooter := mocks.NewMockRebooter(ctrl)
				rebooter.EXPECT().Shutdown().Times(1)
				rebooter.EXPECT().CleanUploadDir().Times(1)
				rebooter.EXPECT().Reset(servers.Reset{
					ImportSubDirs: true,
				}).Times(1)
				rebooter.EXPECT().Boot(gomock.Any()).Times(1)
				return rebooter, ctrl
			},
			mock: func(t *testing.T) *http.Request {
				payload := `{"isd": true}`
				return httptest.NewRequest("POST", "/reset", bytes.NewReader([]byte(payload)))
			},
			handler: func(us uploadServer, w http.ResponseWriter, r *http.Request) {
				us.handleReset(w, r)
			},
			expect: "",
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			rebooter, ctrl := v.rebooterMock(t)
			defer ctrl.Finish()
			us := uploadServer{
				rebooter: rebooter,
			}

			wrt := httptest.NewRecorder()
			req := v.mock(t)
			v.handler(us, wrt, req)
			res, err := ioutil.ReadAll(wrt.Result().Body)

			assert.NoError(t, err)
			assert.Equal(t, v.expect, string(res))
		})
	}
}
