package stub

import (
	"fmt"
	"reflect"
	"sync"
)

var mx = sync.Mutex{}

type stubMapping map[string]map[string][]storage

var stubStorage = stubMapping{}

type storage struct {
	Input  Input
	Output Output
}

func storeStub(stub *Stub) error {
	mx.Lock()
	defer mx.Unlock()

	strg := storage{
		Input:  stub.Input,
		Output: stub.Output,
	}
	if stubStorage[stub.Service] == nil {
		stubStorage[stub.Service] = make(map[string][]storage)
	}
	stubStorage[stub.Service][stub.Method] = append(stubStorage[stub.Service][stub.Method], strg)
	return nil
}

func allStub() stubMapping {
	mx.Lock()
	defer mx.Unlock()
	return stubStorage
}

func findStub(stub *findStubPayload) (*Output, error) {
	mx.Lock()
	defer mx.Unlock()
	if _, ok := stubStorage[stub.Service]; !ok {
		return nil, fmt.Errorf("Can't find stub for Service: %s", stub.Service)
	}

	if _, ok := stubStorage[stub.Service][stub.Method]; !ok {
		return nil, fmt.Errorf("Can't find stub for Service:%s and Method:%s", stub.Service, stub.Method)
	}

	stubs := stubStorage[stub.Service][stub.Method]
	if len(stubs) == 0 {
		return nil, fmt.Errorf("Stub for Service:%s and Method:%s is empty", stub.Service, stub.Method)
	}
	for _, stubrange := range stubs {
		if stubrange.Input.Equals != nil && equals(stub.Data, stubrange.Input.Equals) {
			return &stubrange.Output, nil
		}

		if stubrange.Input.Contains != nil && contains(stubrange.Input.Contains, stub.Data) {
			return &stubrange.Output, nil
		}

		if stubrange.Input.Matches != nil && matches(stubrange.Input.Matches, stub.Data) {
			return &stubrange.Output, nil
		}
	}

	return nil, stubNotFoundError(stub)
}

func stubNotFoundError(stub *findStubPayload) error {
	template := fmt.Sprintf("Can't find stub \n\nService: %s \n\nMethod: %s \n\nInput\n\n", stub.Service, stub.Method)

	template += renderFieldAsString(stub.Data)
	return fmt.Errorf(template)
}

func renderFieldAsString(fields map[string]interface{}) string {
	template := "{\n"
	for key, val := range fields {
		template += fmt.Sprintf("\t%s: %v\n", key, val)
	}
	template += "}"
	return template
}

func equals(input1, input2 map[string]interface{}) bool {
	return reflect.DeepEqual(input1, input2)
}

func contains(expect, actual map[string]interface{}) bool {
	for key, val := range expect {
		actualvalue, ok := actual[key]
		if !ok {
			return ok
		}

		if !reflect.DeepEqual(val, actualvalue) {
			return false
		}
	}
	return true
}

func matches(expect, actual map[string]interface{}) bool {
	for keyExpect, valueExpect := range expect {
		valueExpectMap := valueExpect.(map[string]interface{})
		actualvalue, ok := actual[keyExpect]
		if !ok {
			return ok
		}

		if equals, ok := valueExpectMap["equals"]; ok {
			if !reflect.DeepEqual(equals, actualvalue) {
				return false
			}
		}
	}
	return true
}
