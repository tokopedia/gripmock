package infra

import (
	"encoding/json"
	"fmt"

	"github.com/gripmock/stuber"
)

// Result interface for testing - allows mocking stuber.Result.
type Result interface {
	Found() *stuber.Stub
	Similar() *stuber.Stub
}

// StubNotFoundFormatter provides unified formatting for "stub not found" errors.
type StubNotFoundFormatter struct{}

// NewStubNotFoundFormatter creates a new formatter instance.
func NewStubNotFoundFormatter() *StubNotFoundFormatter {
	return &StubNotFoundFormatter{}
}

// FormatV1 formats error messages for V1 API stub not found scenarios.
func (f *StubNotFoundFormatter) FormatV1(expect stuber.Query, result Result) error {
	template := fmt.Sprintf("Can't find stub \n\nService: %s \n\nMethod: %s \n\nInput\n\n", expect.Service, expect.Method)

	expectString, err := json.MarshalIndent(expect.Data, "", "\t")
	if err != nil {
		template += fmt.Sprintf("Error marshaling input: %v", err)
	} else {
		template += string(expectString)
	}

	if result.Similar() == nil {
		return fmt.Errorf("%s", template) //nolint:err113
	}

	// Add closest matches
	template += f.formatClosestMatches(result)

	return fmt.Errorf("%s", template) //nolint:err113
}

// FormatV2 formats error messages for V2 API stub not found scenarios.
func (f *StubNotFoundFormatter) FormatV2(expect stuber.QueryV2, result Result) error {
	template := fmt.Sprintf("Can't find stub \n\nService: %s \n\nMethod: %s \n\n", expect.Service, expect.Method)

	// Handle streaming input
	template += f.formatInputSection(expect.Input)

	if result.Similar() == nil {
		return fmt.Errorf("%s", template) //nolint:err113
	}

	// Add closest matches
	template += f.formatClosestMatches(result)

	return fmt.Errorf("%s", template) //nolint:err113
}

// formatInputSection formats the input section of the error message.
func (f *StubNotFoundFormatter) formatInputSection(input []map[string]any) string {
	switch {
	case len(input) > 1:
		return f.formatStreamInput(input)
	case len(input) == 1:
		return f.formatSingleInput(input[0])
	default:
		return "Input: (empty)\n\n"
	}
}

// formatStreamInput formats multiple input messages.
func (f *StubNotFoundFormatter) formatStreamInput(input []map[string]any) string {
	template := "Inputs:\n\n"

	for i, inputMsg := range input {
		inputString, err := json.MarshalIndent(inputMsg, "", "\t")
		if err != nil {
			template += fmt.Sprintf("[%d] Error marshaling input: %v\n", i, err)

			continue
		}

		template += fmt.Sprintf("[%d]\n%s\n\n", i, inputString)
	}

	return template
}

// formatSingleInput formats a single input message.
func (f *StubNotFoundFormatter) formatSingleInput(input map[string]any) string {
	template := "Input:\n\n"

	inputString, err := json.MarshalIndent(input, "", "\t")
	if err != nil {
		return template + fmt.Sprintf("Error marshaling input: %v\n\n", err)
	}

	return template + string(inputString) + "\n\n"
}

// formatClosestMatches formats the closest matches section.
func (f *StubNotFoundFormatter) formatClosestMatches(result Result) string {
	addClosestMatch := func(key string, match map[string]any) string {
		if len(match) > 0 {
			matchString, err := json.MarshalIndent(match, "", "\t")
			if err != nil {
				return fmt.Sprintf("\n\nClosest Match \n\n%s: Error marshaling match: %v\nRaw match: %+v", key, err, match)
			}

			return fmt.Sprintf("\n\nClosest Match \n\n%s:%s", key, matchString)
		}

		return ""
	}

	// Check if similar stub has inputs (client streaming)
	similar := result.Similar()
	if similar != nil && similar.IsClientStream() {
		return f.formatStreamClosestMatches(similar, addClosestMatch)
	}

	// Fallback to regular input matching
	var template string

	template += addClosestMatch("equals", result.Similar().Input.Equals)
	template += addClosestMatch("contains", result.Similar().Input.Contains)
	template += addClosestMatch("matches", result.Similar().Input.Matches)

	return template
}

// formatStreamClosestMatches formats closest matches for client streaming inputs.
func (f *StubNotFoundFormatter) formatStreamClosestMatches(stub *stuber.Stub, addClosestMatch func(string, map[string]any) string) string {
	var template string

	for i, inputMsg := range stub.Inputs {
		// Convert InputData to map representation
		inputData := map[string]any{
			"equals":   inputMsg.Equals,
			"contains": inputMsg.Contains,
			"matches":  inputMsg.Matches,
		}
		template += addClosestMatch(fmt.Sprintf("inputs[%d]", i), inputData)
	}

	return template
}
