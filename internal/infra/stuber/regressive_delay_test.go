package stuber_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/gripmock/v3/internal/infra/stuber"
	"github.com/bavix/gripmock/v3/internal/infra/types"
)

func TestOutputEffectiveDelay(t *testing.T) {
	t.Parallel()

	output := stuber.Output{
		Delay:     types.Duration(3 * time.Second),
		DelayType: stuber.DelayTypeRegressive,
		DelayStep: types.Duration(500 * time.Millisecond),
	}

	expected := []time.Duration{
		3 * time.Second,
		2500 * time.Millisecond,
		2 * time.Second,
		1500 * time.Millisecond,
		1 * time.Second,
		500 * time.Millisecond,
		0,
		0,
	}

	for i, delay := range expected {
		require.Equal(t, types.Duration(delay), output.EffectiveDelay(i+1))
	}
}

func TestOutputEffectiveDelayDefault(t *testing.T) {
	t.Parallel()

	for _, delayType := range []stuber.DelayType{"", stuber.DelayTypeDefault} {
		output := stuber.Output{
			Delay:     types.Duration(3 * time.Second),
			DelayType: delayType,
		}

		require.Equal(t, types.Duration(3*time.Second), output.EffectiveDelay(100))
	}
}

func TestRegressiveDelayUsesMatchCountAndPreservesStoredStub(t *testing.T) {
	t.Parallel()

	budgerigar, stub, query := newRegressiveDelayFixture()

	expected := []time.Duration{3 * time.Second, 2500 * time.Millisecond, 2 * time.Second}
	for _, delay := range expected {
		result, err := budgerigar.FindByQuery(query)
		require.NoError(t, err)
		require.NotNil(t, result.Found())
		require.Equal(t, types.Duration(delay), result.Found().Output.Delay)
	}

	stored := budgerigar.FindByID(stub.ID)
	require.Same(t, stub, stored)
	require.Equal(t, types.Duration(3*time.Second), stored.Output.Delay)
}

func TestRegressiveDelayIsIsolatedBySession(t *testing.T) {
	t.Parallel()

	budgerigar, _, query := newRegressiveDelayFixture()

	sessions := []string{"A", "B", "A", ""}
	expected := []time.Duration{3 * time.Second, 3 * time.Second, 2500 * time.Millisecond, 3 * time.Second}

	for i, session := range sessions {
		query.Session = session
		result, err := budgerigar.FindByQuery(query)
		require.NoError(t, err)
		require.Equal(t, types.Duration(expected[i]), result.Found().Output.Delay)
	}
}

func TestRegressiveDelayResetsAfterStubDeletion(t *testing.T) {
	t.Parallel()

	budgerigar, stub, query := newRegressiveDelayFixture()

	for _, expected := range []time.Duration{3 * time.Second, 2500 * time.Millisecond} {
		result, err := budgerigar.FindByQuery(query)
		require.NoError(t, err)
		require.Equal(t, types.Duration(expected), result.Found().Output.Delay)
	}

	require.Equal(t, 1, budgerigar.DeleteByID(stub.ID))
	budgerigar.PutMany(stub)

	result, err := budgerigar.FindByQuery(query)
	require.NoError(t, err)
	require.Equal(t, types.Duration(3*time.Second), result.Found().Output.Delay)
}

func TestRegressiveDelayConcurrentMatchesUseUniqueOrdinals(t *testing.T) {
	t.Parallel()

	budgerigar, query := newConcurrentRegressiveDelayFixture()

	const concurrency = 50

	type concurrentResult struct {
		delay types.Duration
		err   error
	}

	results := make(chan concurrentResult, concurrency)

	var group sync.WaitGroup

	group.Add(concurrency)

	for range concurrency {
		go func() {
			defer group.Done()

			result, err := budgerigar.FindByQuery(query)
			if err != nil {
				results <- concurrentResult{err: err}

				return
			}

			results <- concurrentResult{delay: result.Found().Output.Delay}
		}()
	}

	group.Wait()
	close(results)

	actual := make(map[types.Duration]struct{}, concurrency)

	for result := range results {
		require.NoError(t, result.err)
		actual[result.delay] = struct{}{}
	}

	require.Len(t, actual, concurrency)

	for i := range concurrency {
		expected := types.Duration(time.Duration(concurrency-i) * time.Millisecond)
		_, found := actual[expected]
		require.True(t, found, "missing effective delay %s", time.Duration(expected))
	}
}

func newConcurrentRegressiveDelayFixture() (*stuber.Budgerigar, stuber.Query) {
	budgerigar := stuber.NewBudgerigar()
	budgerigar.PutMany(&stuber.Stub{
		Service: "example.Service",
		Method:  "Get",
		Input:   stuber.InputData{Equals: map[string]any{"key": "value"}},
		Output: stuber.Output{
			Data:      map[string]any{"result": "ok"},
			Delay:     types.Duration(50 * time.Millisecond),
			DelayType: stuber.DelayTypeRegressive,
			DelayStep: types.Duration(time.Millisecond),
		},
	})

	query := stuber.Query{
		Service: "example.Service",
		Method:  "Get",
		Input:   []map[string]any{{"key": "value"}},
	}

	return budgerigar, query
}

func TestRegressiveDelayAdvancesForBidiMatches(t *testing.T) {
	t.Parallel()

	budgerigar := stuber.NewBudgerigar()
	budgerigar.PutMany(&stuber.Stub{
		Service: "example.Chat",
		Method:  "Talk",
		Inputs: []stuber.InputData{
			{Equals: map[string]any{"message": "one"}},
			{Equals: map[string]any{"message": "two"}},
			{Equals: map[string]any{"message": "three"}},
		},
		Output: stuber.Output{
			Stream:    []any{map[string]any{"message": "ok"}},
			Delay:     types.Duration(30 * time.Millisecond),
			DelayType: stuber.DelayTypeRegressive,
			DelayStep: types.Duration(10 * time.Millisecond),
		},
	})

	result, err := budgerigar.FindByQueryBidi(stuber.QueryBidi{Service: "example.Chat", Method: "Talk"})
	require.NoError(t, err)

	messages := []string{"one", "two", "three"}
	for i, expected := range []time.Duration{30 * time.Millisecond, 20 * time.Millisecond, 10 * time.Millisecond} {
		found, nextErr := result.Next(map[string]any{"message": messages[i]})
		require.NoError(t, nextErr, "message %d", i)
		require.Equal(t, types.Duration(expected), found.Output.Delay)
	}
}

func newRegressiveDelayFixture() (*stuber.Budgerigar, *stuber.Stub, stuber.Query) {
	budgerigar := stuber.NewBudgerigar()
	stub := &stuber.Stub{
		Service: "example.Service",
		Method:  "Get",
		Input:   stuber.InputData{Equals: map[string]any{"key": "value"}},
		Output: stuber.Output{
			Data:      map[string]any{"result": "ok"},
			Delay:     types.Duration(3 * time.Second),
			DelayType: stuber.DelayTypeRegressive,
			DelayStep: types.Duration(500 * time.Millisecond),
		},
	}
	budgerigar.PutMany(stub)

	query := stuber.Query{
		Service: "example.Service",
		Method:  "Get",
		Input:   []map[string]any{{"key": "value"}},
	}

	return budgerigar, stub, query
}
