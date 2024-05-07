package stub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/icrowley/fake"
	"google.golang.org/grpc/codes"

	"github.com/go-chi/chi"
)

type Options struct {
	Port     string
	BindAddr string
	StubPath string
}

const DEFAULT_PORT = "4771"

func RunStubServer(opt Options) {
	if opt.Port == "" {
		opt.Port = DEFAULT_PORT
	}
	addr := opt.BindAddr + ":" + opt.Port
	r := chi.NewRouter()
	r.Post("/add", addStub)
	r.Get("/", listStub)
	r.Post("/find", handleFindStub)
	r.Get("/clear", handleClearStub)

	if opt.StubPath != "" {
		readStubFromFile(opt.StubPath)
	}

	fmt.Println("Serving stub admin on http://" + addr)
	go func() {
		err := http.ListenAndServe(addr, r)
		log.Fatal(err)
	}()
}

func responseError(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

type Stub struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Input   Input  `json:"input"`
	Output  Output `json:"output"`
	Order   int    `json:"order"`
}

type Input struct {
	Equals   map[string]interface{} `json:"equals"`
	Contains map[string]interface{} `json:"contains"`
	Matches  map[string]interface{} `json:"matches"`
}

type Output struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
	Code  *codes.Code            `json:"code,omitempty"`
}

func addStub(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(err, w)
		return
	}

	stub := new(Stub)
	err = json.Unmarshal(body, stub)
	if err != nil {
		responseError(err, w)
		return
	}

	err = validateStub(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	err = storeStub(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	w.Write([]byte("Success add stub"))
}

func listStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allStub())
}

func validateStub(stub *Stub) error {
	if stub.Service == "" {
		return fmt.Errorf("Service name can't be empty")
	}

	if stub.Method == "" {
		return fmt.Errorf("Method name can't be emtpy")
	}

	// due to golang implementation
	// method name must capital
	stub.Method = strings.Title(stub.Method)

	switch {
	case stub.Input.Contains != nil:
		break
	case stub.Input.Equals != nil:
		break
	case stub.Input.Matches != nil:
		break
	default:
		return fmt.Errorf("Input cannot be empty")
	}

	// TODO: validate all input case

	if stub.Output.Error == "" && stub.Output.Data == nil && stub.Output.Code == nil {
		return fmt.Errorf("Output can't be empty")
	}
	return nil
}

type findStubPayload struct {
	Service string                 `json:"service"`
	Method  string                 `json:"method"`
	Data    map[string]interface{} `json:"data"`
}

func handleFindStub(w http.ResponseWriter, r *http.Request) {
	stub := new(findStubPayload)
	err := json.NewDecoder(r.Body).Decode(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	// due to golang implementation
	// method name must capital
	stub.Method = strings.Title(stub.Method)

	outputP, err := findStub(stub)

	if err != nil {
		log.Println(err)
		responseError(err, w)
		return
	}

	output := echoInputData(outputP, stub.Data)
	fillFakeData(&output)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&output)
}

func handleClearStub(w http.ResponseWriter, r *http.Request) {
	clearStorage()
	w.Write([]byte("OK"))
}

func echoInputData(outputP *Output, data map[string]interface{}) Output {
	output := *outputP
	outputData := make(map[string]interface{})
	for key, value := range outputP.Data {
		outputData[key] = value
	}

	for outputKey, outputVal := range outputData {
		if reflect.TypeOf(outputVal).String() == "string" {
			outputValString := outputVal.(string)
			if strings.Contains(outputValString, "input.") {
				for dataK, dataV := range data {
					if strings.Contains(outputValString, "{{input."+dataK+"}}") {
						newResponse := strings.Replace(outputValString, "{{input."+dataK+"}}", toString(dataV), -1)
						outputData[outputKey] = newResponse
					}
				}
			}
		}
	}
	output.Data = outputData

	return output
}

func fillFakeData(output *Output) {
	for outputDataK, outputDataV := range output.Data {
		if reflect.TypeOf(outputDataV).String() == "string" {
			data := outputDataV.(string)
			output.Data[outputDataK] = data
			if strings.Contains(data, "fake.") {
				re := regexp.MustCompile(`{{fake.([0-9A-Za-z()]+)}}`)
				matches := re.FindAllStringSubmatch(data, -1)

				for _, match := range matches {
					if len(match) > 1 && strings.Contains(match[1], "DigitsN") {
						output.Data[outputDataK] = strings.Replace(output.Data[outputDataK].(string), match[0], fakeDigitsN(match[1]), 1)
					} else if len(match) > 1 && strings.Contains(match[1], "Digits") {
						output.Data[outputDataK] = strings.Replace(output.Data[outputDataK].(string), match[0], fakeDigits(), 1)
					} else {
						output.Data[outputDataK] = "Unsupported fake method"
					}
				}
			}
		}
	}
}

func fakeDigits() string {
	return fake.Digits()
}

func fakeDigitsN(sig string) string {
	re := regexp.MustCompile(`DigitsN\((\d+)\)`)
	match := re.FindStringSubmatch(sig)
	if len(match) > 1 {
		n, err := strconv.Atoi(match[1])
		if err != nil {
			return "error"
		}
		return fake.DigitsN(n)
	}
	return "error"
}

func toString[T any](value T) string {
	// Get the reflect type of the value
	valType := reflect.TypeOf(value)

	// Check if the value is already a string
	if valType.Kind() == reflect.String {
		return fmt.Sprint(value) // Return the string value as is
	}

	return "\"" + fmt.Sprint(value) + "\"" // Add quotes for non-string values
}
