package upload

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/quintans/gripmock/servers"
)

type Options struct {
	Port     string
	BindAddr string
}

const defaultUploadPort = "4772"

func RunUploadServer(ctx context.Context, opt Options, rebooter servers.Rebooter) <-chan struct{} {
	us := uploadServer{
		rebooter: rebooter,
		ctx:      ctx,
	}
	if opt.Port == "" {
		opt.Port = defaultUploadPort
	}
	addr := opt.BindAddr + ":" + opt.Port
	r := chi.NewRouter()
	r.Post("/upload", us.handleUpload)
	r.Post("/reset", us.handleReset)
	r.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I'm alive: " + servers.Version))
	})
	fs := http.FileServer(http.Dir("/"))
	http.StripPrefix("/dir/", fs)
	r.Get("/dir/*", http.StripPrefix("/dir/", fs).ServeHTTP)

	log.Println("Serving proto upload on http://" + addr)
	srv := http.Server{
		Addr:    addr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("HTTP proto upload  server ListenAndServe: %v", err)
		}
	}()
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Printf("HTTP proto upload server Shutdown")
		if err := srv.Shutdown(ctx); err != nil && err != context.Canceled {
			// Error from closing listeners, or context cancel:
			log.Printf("Error: HTTP proto upload server Shutdown: %v", err)
		}
		close(done)
	}()
	return done
}

func responseError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

// Unzip will decompress a zip bytes, moving all files and folders
// within the zip bytes (parameter 1) to an output directory (parameter 2).
func Unzip(src []byte, dest string) error {
	log.Println("Unzipping protos to", dest)

	r, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		return err
	}

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fPath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fPath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

type uploadServer struct {
	rebooter servers.Rebooter
	ctx      context.Context
}

func (s uploadServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	log.Println("Receiving new proto files.")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error Retrieving the File: %v\n", err)
		responseError(w, err)
		return
	}

	// shutdown
	s.rebooter.Shutdown()

	err = Unzip(body, s.rebooter.UploadDir())
	if err != nil {
		responseError(w, err)
		return
	}

	// boot
	err = s.rebooter.Boot(s.ctx)
	if err != nil {
		responseError(w, err)
	}
}

func (s uploadServer) handleReset(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err)
		return
	}

	reset := servers.Reset{}
	err = json.Unmarshal(body, &reset)
	if err != nil {
		responseError(w, err)
		return
	}

	log.Printf("Resetting %+v\n", reset)

	// shutdown
	s.rebooter.Shutdown()

	err = s.rebooter.CleanUploadDir()
	if err != nil {
		log.Fatalf("Unable to clean upload dir: %v", err)
	}

	s.rebooter.Reset(reset)

	// boot
	err = s.rebooter.Boot(s.ctx)
	if err != nil {
		responseError(w, err)
	}
}
