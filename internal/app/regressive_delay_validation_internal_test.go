package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/gripmock/v3/internal/infra/stuber"
	"github.com/bavix/gripmock/v3/internal/infra/types"
)

func TestRegressiveDelayValidation(t *testing.T) {
	t.Parallel()

	validator, err := NewStubValidator()
	require.NoError(t, err)

	tests := []struct {
		name      string
		delay     time.Duration
		delayType stuber.DelayType
		step      time.Duration
		valid     bool
	}{
		{name: "omitted type", delay: time.Second, valid: true},
		{name: "explicit default", delay: time.Second, delayType: stuber.DelayTypeDefault, valid: true},
		{name: "regressive", delay: time.Second, delayType: stuber.DelayTypeRegressive, step: 100 * time.Millisecond, valid: true},
		{name: "unknown type", delay: time.Second, delayType: "random", step: 100 * time.Millisecond},
		{name: "missing base delay", delayType: stuber.DelayTypeRegressive, step: 100 * time.Millisecond},
		{name: "negative base delay", delay: -time.Second, delayType: stuber.DelayTypeRegressive, step: 100 * time.Millisecond},
		{name: "missing step", delay: time.Second, delayType: stuber.DelayTypeRegressive},
		{name: "negative step", delay: time.Second, delayType: stuber.DelayTypeRegressive, step: -100 * time.Millisecond},
		{name: "step with omitted type", delay: time.Second, step: 100 * time.Millisecond},
		{name: "step with default type", delay: time.Second, delayType: stuber.DelayTypeDefault, step: 100 * time.Millisecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			stub := &stuber.Stub{
				Service: "example.Service",
				Method:  "Get",
				Input:   stuber.InputData{Contains: map[string]any{}},
				Output: stuber.Output{
					Data:      map[string]any{"result": "ok"},
					Delay:     types.Duration(test.delay),
					DelayType: test.delayType,
					DelayStep: types.Duration(test.step),
				},
			}

			validationErr := validator.Struct(stub)
			if test.valid {
				require.NoError(t, validationErr)
			} else {
				require.Error(t, validationErr)
			}
		})
	}
}
