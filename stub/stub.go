package stub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"

	"github.com/go-chi/chi/v5"
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
	if _, err = w.Write([]byte(err.Error())); err != nil {
		log.Println("Error writing response: %w", err)
	}
}

type Stub struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Input   Input  `json:"input"`
	Output  Output `json:"output"`
}

type Input struct {
	Equals          map[string]interface{} `json:"equals"`
	EqualsUnordered map[string]interface{} `json:"equals_unordered"`
	Contains        map[string]interface{} `json:"contains"`
	Matches         map[string]interface{} `json:"matches"`
}

type Output struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
	Code  *codes.Code            `json:"code,omitempty"`
}

func addStub(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
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

	if _, err = w.Write([]byte("Success add stub")); err != nil {
		log.Println("Error writing response: %w", err)
	}
}

func listStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allStub()); err != nil {
		log.Println("Error writing listStub response: %w", err)
	}
}

func validateStub(stub *Stub) error {
	if stub.Service == "" {
		return fmt.Errorf("service name can't be empty")
	}

	if stub.Method == "" {
		return fmt.Errorf("method name can't be emtpy")
	}

	// due to golang implementation
	// method name must capital
	stub.Method = cases.Title(language.Und, cases.NoLower).String(stub.Method)

	switch {
	case stub.Input.Contains != nil:
		break
	case stub.Input.Equals != nil:
		break
	case stub.Input.EqualsUnordered != nil:
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
	stub.Method = cases.Title(language.Und, cases.NoLower).String(stub.Method)

	output, err := findStub(stub)
	if err != nil {
		log.Println(err)
		responseError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output); err != nil {
		log.Println("Error writing handleFindStub response: %w", err)
	}
}

func handleClearStub(w http.ResponseWriter, r *http.Request) {
	clearStorage()
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Println("Error writing handleClearStub response: %w", err)
	}
}
