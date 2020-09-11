package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type Options struct {
	Port     string
	BindAddr string
	StubPath string
}

const DEFAULT_PORT = "4771"

func RunStubServer(ctx context.Context, opt Options) <-chan struct{} {
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

	log.Println("Serving stub admin on http://" + addr)
	srv := http.Server{
		Addr:    addr,
		Handler: r,
	}
	done := make(chan struct{})
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("HTTP stub server ListenAndServe: %v", err)
		}
		close(done)
	}()
	go func() {
		<-ctx.Done()
		log.Printf("HTTP stub server Shutdown")
		if err := srv.Shutdown(ctx); err != nil && err != context.Canceled {
			// Error from closing listeners, or context cancel:
			log.Printf("Error: HTTP stub server Shutdown: %v", err)
		}
	}()

	return done
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

	stub.Method = strings.Title(stub.Method)

	output, err := findStub(stub)
	if err != nil {
		log.Println(err)
		responseError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func handleClearStub(w http.ResponseWriter, r *http.Request) {
	clearStorage()
	w.Write([]byte("OK"))
}
