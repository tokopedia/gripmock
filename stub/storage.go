package stub

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/renstrom/fuzzysearch/fuzzy"
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

type closeMatch struct {
	rule   string
	expect map[string]interface{}
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

	closestMatch := []closeMatch{}
	for _, stubrange := range stubs {
		if expect := stubrange.Input.Equals; expect != nil {
			closestMatch = append(closestMatch, closeMatch{"equals", expect})
			if equals(stub.Data, expect) {
				return &stubrange.Output, nil
			}
		}

		if expect := stubrange.Input.Contains; expect != nil {
			closestMatch = append(closestMatch, closeMatch{"contains", expect})
			if contains(stubrange.Input.Contains, stub.Data) {
				return &stubrange.Output, nil
			}
		}

		if expect := stubrange.Input.Matches; expect != nil {
			closestMatch = append(closestMatch, closeMatch{"matches", expect})
			if matches(stubrange.Input.Matches, stub.Data) {
				return &stubrange.Output, nil
			}
		}
	}

	return nil, stubNotFoundError(stub, closestMatch)
}

func stubNotFoundError(stub *findStubPayload, closestMatches []closeMatch) error {
	template := fmt.Sprintf("Can't find stub \n\nService: %s \n\nMethod: %s \n\nInput\n\n", stub.Service, stub.Method)
	expectString := renderFieldAsString(stub.Data)
	template += expectString

	lowestFuzzyRank := struct {
		rank  int
		match closeMatch
	}{-1, closeMatch{}}
	for _, closeMatchValue := range closestMatches {
		closeMatchString := renderFieldAsString(closeMatchValue.expect)
		src := expectString
		tgt := closeMatchString
		// swap it
		if len(src) > len(tgt) {
			tmp := tgt
			tgt = src
			src = tmp
		}
		rank := fuzzy.RankMatch(src, tgt)
		if rank == -1 {
			continue
		}
		// 0 is the closest match
		if rank < lowestFuzzyRank.rank || lowestFuzzyRank.rank == -1 {
			lowestFuzzyRank.rank = rank
			lowestFuzzyRank.match = closeMatchValue
		}
	}

	var closestMatch closeMatch
	if lowestFuzzyRank.rank == -1 {
		closestMatch = closestMatches[0]
	} else {
		closestMatch = lowestFuzzyRank.match
	}

	closestMatchString := renderFieldAsString(closestMatch.expect)
	template += fmt.Sprintf("\n\nClosest Match \n\n%s:%s", closestMatch.rule, closestMatchString)

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
