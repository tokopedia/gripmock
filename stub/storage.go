package stub

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var mx = sync.Mutex{}

// below represent map[servicename][methodname][]expectations
type stubMapping map[string]map[string][]storage

type matchFunc func(interface{}, interface{}) bool

var stubStorage = stubMapping{}

type storage struct {
	Input  Input
	Output Output
}

func storeStub(stub *Stub) error {
	return stubStorage.storeStub(stub)
}

func (sm *stubMapping) storeStub(stub *Stub) error {
	mx.Lock()
	defer mx.Unlock()

	strg := storage{
		Input:  stub.Input,
		Output: stub.Output,
	}
	if (*sm)[stub.Service] == nil {
		(*sm)[stub.Service] = make(map[string][]storage)
	}
	(*sm)[stub.Service][stub.Method] = append((*sm)[stub.Service][stub.Method], strg)
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
		return nil, fmt.Errorf("can't find stub for Service: %s", stub.Service)
	}

	if _, ok := stubStorage[stub.Service][stub.Method]; !ok {
		return nil, fmt.Errorf("can't find stub for Service:%s and Method:%s", stub.Service, stub.Method)
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

		if expect := stubrange.Input.EqualsUnordered; expect != nil {
			closestMatch = append(closestMatch, closeMatch{"equals_unordered", expect})
			if equalsUnordered(stub.Data, expect) {
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

	if len(closestMatches) == 0 {
		return fmt.Errorf(template)
	}

	highestRank := struct {
		rank  float32
		match closeMatch
	}{0, closeMatch{}}
	for _, closeMatchValue := range closestMatches {
		rank := rankMatch(expectString, closeMatchValue.expect)

		// the higher the better
		if rank > highestRank.rank {
			highestRank.rank = rank
			highestRank.match = closeMatchValue
		}
	}

	var closestMatch closeMatch
	if highestRank.rank == 0 {
		closestMatch = closestMatches[0]
	} else {
		closestMatch = highestRank.match
	}

	closestMatchString := renderFieldAsString(closestMatch.expect)
	template += fmt.Sprintf("\n\nClosest Match \n\n%s:%s", closestMatch.rule, closestMatchString)

	return fmt.Errorf(template)
}

// we made our own simple ranking logic
// count the matches field_name and value then compare it with total field names and values
// the higher the better
func rankMatch(expect string, closeMatch map[string]interface{}) float32 {
	occurence := 0
	for key, value := range closeMatch {
		if fuzzy.Match(key+":", expect) {
			occurence++
		}

		if fuzzy.Match(fmt.Sprint(value), expect) {
			occurence++
		}
	}

	if occurence == 0 {
		return 0
	}
	totalFields := len(closeMatch) * 2
	return float32(occurence) / float32(totalFields)
}

func renderFieldAsString(fields map[string]interface{}) string {
	template := "{\n"
	for key, val := range fields {
		template += fmt.Sprintf("\t%s: %v\n", key, val)
	}
	template += "}"
	return template
}

func deepEqual(expect, actual interface{}) bool {
	return reflect.DeepEqual(expect, actual)
}

func regexMatch(expect, actual interface{}) bool {
	expectedStr, expectedStringOk := expect.(string)
	actualStr, actualStringOk := actual.(string)

	if expectedStringOk && actualStringOk {
		match, err := regexp.Match(expectedStr, []byte(actualStr))
		if err != nil {
			log.Printf("Error on matching regex %s with %s error:%v\n", expect, actual, err)
		}
		return match
	}

	return deepEqual(expect, actual)
}

func equals(expect, actual map[string]interface{}) bool {
	return find(expect, actual, true, true, deepEqual, false)
}

func equalsUnordered(expect, actual map[string]interface{}) bool {
	return find(expect, actual, true, true, deepEqual, true)
}

func contains(expect, actual map[string]interface{}) bool {
	return find(expect, actual, true, false, deepEqual, false)
}

func matches(expect, actual map[string]interface{}) bool {
	return find(expect, actual, true, false, regexMatch, false)
}

func equalsIgnoreOrder(expect, actual interface{}) bool {
	expectSlice, expectOk := expect.([]interface{})
	actualSlice, actualOk := actual.([]interface{})
	if !expectOk || !actualOk {
		return false
	}
	if len(expectSlice) != len(actualSlice) {
		return false
	}
	sort.Slice(expectSlice, func(i, j int) bool {
		return fmt.Sprint(expectSlice[i]) < fmt.Sprint(expectSlice[j])
	})
	sort.Slice(actualSlice, func(i, j int) bool {
		return fmt.Sprint(actualSlice[i]) < fmt.Sprint(actualSlice[j])
	})
	return reflect.DeepEqual(expectSlice, actualSlice)
}

func find(expect, actual interface{}, acc, exactMatch bool, f matchFunc, ignoreOrder bool) bool {

	// circuit brake
	if !acc {
		return false
	}

	expectArrayValue, expectArrayOk := expect.([]interface{})
	if expectArrayOk {

		actualArrayValue, actualArrayOk := actual.([]interface{})
		if !actualArrayOk {
			acc = false
			return acc
		}

		if exactMatch {
			if len(expectArrayValue) != len(actualArrayValue) {
				acc = false
				return acc
			}
		} else {
			if len(expectArrayValue) > len(actualArrayValue) {
				acc = false
				return acc
			}
		}

		if expectArrayOk && actualArrayOk && ignoreOrder {
			return equalsIgnoreOrder(expectArrayValue, actualArrayValue)
		}

		for expectItemIndex, expectItemValue := range expectArrayValue {
			actualItemValue := actualArrayValue[expectItemIndex]
			acc = find(expectItemValue, actualItemValue, acc, exactMatch, f, ignoreOrder)
		}

		return acc
	}

	expectMapValue, expectMapOk := expect.(map[string]interface{})
	if expectMapOk {

		actualMapValue, actualMapOk := actual.(map[string]interface{})
		if !actualMapOk {
			acc = false
			return acc
		}

		if exactMatch {
			if len(expectMapValue) != len(actualMapValue) {
				acc = false
				return acc
			}
		} else {
			if len(expectMapValue) > len(actualMapValue) {
				acc = false
				return acc
			}
		}

		for expectItemKey, expectItemValue := range expectMapValue {
			actualItemValue := actualMapValue[expectItemKey]
			acc = find(expectItemValue, actualItemValue, acc, exactMatch, f, ignoreOrder)
		}

		return acc
	}

	return f(expect, actual)
}

func clearStorage() {
	mx.Lock()
	defer mx.Unlock()

	stubStorage = stubMapping{}
}

func readStubFromFile(path string) int {
	return stubStorage.readStubFromFile(path)
}

func (sm *stubMapping) readStubFromFile(path string) int {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Can't read stub from %s. %v\n", path, err)
		return 0
	}

	count := 0
	for _, file := range files {
		if file.IsDir() {
			count += sm.readStubFromFile(path + "/" + file.Name())
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".json") {
			continue
		}

		filePath := path + "/" + file.Name()
		byt, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error when reading file %s. %v. skipping...", file.Name(), err)
			continue
		}

		// Try to unmarshal as array first
		var stubs []*Stub
		err = json.Unmarshal(byt, &stubs)
		if err == nil && len(stubs) > 0 {
			// Successfully unmarshaled as array
			log.Printf("Successfully unmarshaled %s as array with %d stubs", file.Name(), len(stubs))
			for _, s := range stubs {
				if err = sm.storeStub(s); err != nil {
					log.Printf("Error when storing Stub from %s. %v. skipping...", file.Name(), err)
				} else {
					count++
				}
			}
			continue
		}

		// If array unmarshal failed, try as single stub
		var stub Stub
		err = json.Unmarshal(byt, &stub)
		if err != nil {
			log.Printf("Error when unmarshalling file %s. %v. skipping...", file.Name(), err)
			continue
		}

		if err = sm.storeStub(&stub); err != nil {
			log.Printf("Error when storing Stub from %s. %v. skipping...", file.Name(), err)
		} else {
			count++
		}
	}
	
	return count
}
