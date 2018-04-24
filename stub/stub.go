package stub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func RunStubServer(port string) {
	r := chi.NewRouter()
	r.Post("/add", addStub)
	r.Get("/", listStub)
	r.Post("/find", handleFindStub)
	fmt.Println("Serving HTTP server on localhost" + port)
	go func() {
		err := http.ListenAndServe(port, r)
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
}

type Input struct {
	Equals   map[string]interface{} `json:"equals"`
	Contains map[string]interface{} `json:"contains"`
	Matches  map[string]interface{} `json:"matches"`
}

type Output struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
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

	success := struct {
		Success string `json:"success"`
	}{
		"Success add stub",
	}
	json.NewEncoder(w).Encode(success)
}

func listStub(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(allStub())
}

func validateStub(stub *Stub) error {
	if stub.Service == "" {
		return fmt.Errorf("Service name can't be empty")
	}

	if stub.Method == "" {
		return fmt.Errorf("Method name can't be emtpy")
	}

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

	if stub.Output.Error == "" && stub.Output.Data == nil {
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

	output, err := findStub(stub)
	if err != nil {
		log.Println(err)
		responseError(err, w)
		return
	}

	json.NewEncoder(w).Encode(output)
}
