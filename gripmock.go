package main

import (
	"archive/zip"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/tokopedia/gripmock/stub"
)

const (
	version    = "v0.0.1"
	binaryName = "grpcserver"
)

var (
	upProtoFolder = "/temp/up-proto"
	goPath        = ""
)

func main() {
	ver := flag.Bool("v", false, "returns the version")
	output := flag.String("o", "", "directory to output server.go. Default is $GOPATH/src/grpc/")
	grpcPort := flag.String("grpc-port", "4770", "Port of gRPC tcp server")
	grpcBindAddr := flag.String("grpc-listen", "", "Address the gRPC server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	adminport := flag.String("admin-port", "4771", "Port of stub admin server")
	adminBindAddr := flag.String("admin-listen", "", "Address the admin server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	stubPath := flag.String("stub", "", "Path where the stub files are (Optional)")
	upPort := flag.String("up-port", "4772", "Port of upload proto server")
	upBindAddr := flag.String("up-listen", "", "Address the upload proto server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	imports := flag.String("imports", "", "comma separated imports path. Path /protobuf is always set. It is where gripmock Dockerfile install WKT protos")
	importSubdirs := flag.Bool("isd", false, "Immediate sub dirs of the upload, will be imported")
	// for backwards compatibility
	if len(os.Args) > 1 && os.Args[1] == "gripmock" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	if *ver {
		fmt.Println("version:", version)
		return
	}

	flag.Parse()
	log.Println("Starting GripMock", version)
	goPath = os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("GOPATH is empty")
	}
	if *output == "" {
		*output = os.Getenv("GOPATH") + "/src/grpc"
	}
	upProtoFolder = goPath + upProtoFolder
	// for safety
	*output += "/"
	if _, err := os.Stat(*output); os.IsNotExist(err) {
		os.Mkdir(*output, os.ModePerm)
	}

	err := os.RemoveAll(upProtoFolder)
	if err != nil {
		log.Fatalf("did not remove %s: %v", upProtoFolder, err)
	}
	err = os.MkdirAll(upProtoFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("did not create %s: %v", upProtoFolder, err)
	}

	importDirs := []string{"/protobuf"}
	impSplit := strings.Split(*imports, ",")
	if len(*imports) > 0 {
		importDirs = append(importDirs, impSplit...)
	}

	// parse proto files
	args := flag.Args()
	protoPaths, imps := expandDirs(args)

	// if the first arg is a file, add the its dir to the list of imports
	if len(args) > 0 {
		name := args[0]
		fi, err := os.Stat(name)
		if err != nil {
			log.Fatalf("Unable to get the status of %s: %v", name, err)
			return
		}
		if fi.Mode().IsRegular() {
			protodirs := strings.Split(name, "/")
			protodir := ""
			if len(protodirs) > 0 {
				protodir = strings.Join(protodirs[:len(protodirs)-1], "/")
			}
			imps = append(imps, protodir)
		}
	}

	// only add argument folders if they are not inside imports list
	for _, i := range imps {
		found := false
		for _, is := range impSplit {
			if strings.Contains(is, i) {
				found = true
			}
		}
		if !found {
			importDirs = append(importDirs, i)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	srvrs := &servers{
		options: stub.Options{
			StubPath: *stubPath,
			Port:     *adminport,
			BindAddr: *adminBindAddr,
		},
		params: protocParam{
			protoPaths:    protoPaths,
			adminPort:     *adminport,
			grpcAddress:   *grpcBindAddr,
			grpcPort:      *grpcPort,
			imports:       importDirs,
			importSubdirs: *importSubdirs,
		},
		grpcOutput: *output,
	}

	err = srvrs.boot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	upDone := uploadServer(ctx, Options{
		BindAddr: *upBindAddr,
		Port:     *upPort,
		Output:   upProtoFolder,
	}, srvrs)

	var term = make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case <-term:
		log.Println("Signaled to shutdown")
		cancel()
		srvrs.shutdown()
		<-upDone
	}
}

type servers struct {
	options    stub.Options
	params     protocParam
	grpcOutput string
	cancel     context.CancelFunc
	stubDone   <-chan struct{}
	grpcDone   <-chan struct{}
}

func (s *servers) boot(ctx context.Context) error {
	protos, _ := expandDirs([]string{upProtoFolder})
	protoPaths := append(s.params.protoPaths, protos...)
	if len(protoPaths) == 0 {
		log.Println("No proto files found. Skipping stub and grpc boot.")
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// clean output folder
	err := os.RemoveAll(s.grpcOutput)
	if err != nil {
		return err
	}
	err = os.Mkdir(s.grpcOutput, os.ModePerm)
	if err != nil {
		return err
	}

	// run admin stub server
	s.stubDone = stub.RunStubServer(ctx, s.options)

	var importDirs []string
	if s.params.importSubdirs {
		importDirs = append(s.params.imports, importSudDirs([]string{upProtoFolder})...)
	} else {
		importDirs = append(s.params.imports, upProtoFolder)
	}

	pp := protocParam{
		imports:     importDirs,
		adminPort:   s.params.adminPort,
		grpcAddress: s.params.grpcAddress,
		grpcPort:    s.params.grpcPort,
		protoPaths:  protoPaths,
	}

	// generate pb.go and grpc server based on proto
	err = generateProtoc(pp, s.grpcOutput)
	if err != nil {
		return err
	}

	// build the server
	err = buildServer(s.grpcOutput)
	if err != nil {
		return err
	}

	// and run
	s.grpcDone, err = runGrpcServer(ctx, s.grpcOutput)
	return err
}

func (s *servers) shutdown() {
	if s.cancel != nil {
		s.cancel()
		<-s.stubDone
		<-s.grpcDone
	}
}

func getProtoName(path string) string {
	paths := strings.Split(path, "/")
	filename := paths[len(paths)-1]
	return strings.Split(filename, ".")[0]
}

type protocParam struct {
	protoPaths    []string
	adminPort     string
	grpcAddress   string
	grpcPort      string
	imports       []string
	importSubdirs bool
}

func generateProtoc(param protocParam, output string) error {
	src := goPath + "/src"
	args := []string{}
	// include well-known-types
	for _, i := range param.imports {
		args = append(args, "-I", i)
		log.Println("Importing", i)
	}
	args = append(args, "--go_out=plugins=grpc:"+src)
	args = append(args, fmt.Sprintf("--gripmock_out=admin-port=%s,grpc-address=%s,grpc-port=%s:%s",
		param.adminPort, param.grpcAddress, param.grpcPort, output))
	args = append(args, param.protoPaths...)
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	err := protoc.Run()
	if err != nil {
		return fmt.Errorf("Fail on protoc: %w", err)
	}

	// change package to "main" on top level generated code
	files, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatalf("Can't read dir for imports from %s. %v\n", output, err)
	}

	for _, fi := range files {
		name := fi.Name()
		if fi.Mode().IsRegular() && strings.HasSuffix(name, ".pb.go") {
			source := filepath.Join(src, name)
			sed := exec.Command("sed", "-i", `s/^package \w*$/package main/`, source)
			sed.Stderr = os.Stderr
			sed.Stdout = os.Stdout
			err = sed.Run()
			if err != nil {
				return fmt.Errorf("Fail on sed: %w", err)
			}
			target := filepath.Join(output, name)
			err = os.Remove(target)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("Fail to remove target go file after sed: %w", err)
			}
			err = os.Rename(source, target)
			if err != nil {
				return fmt.Errorf("Fail to rename after sed: %w", err)
			}
		}
	}

	return nil
}

func buildServer(output string) error {
	files, err := ioutil.ReadDir(output)
	if err != nil {
		log.Fatalf("Can't read dir for go files %s. %v\n", output, err)
	}

	args := []string{"build", "-o", binaryName}
	for _, fi := range files {
		name := fi.Name()
		if fi.Mode().IsRegular() && strings.HasSuffix(name, ".go") {
			args = append(args, name)
		}
	}

	build := exec.Command("go", args...)
	build.Dir = output
	build.Env = []string{
		"GO111MODULE=off",
		"GOPATH=" + goPath,
		"HOME=" + os.Getenv("HOME"),
	}
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err = build.Run()
	return err
}

func runGrpcServer(ctx context.Context, output string) (<-chan struct{}, error) {
	run := exec.Command(output + binaryName)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	err := run.Start()
	if err != nil {
		return nil, err
	}
	log.Printf("grpc server pid: %d\n", run.Process.Pid)
	done := make(chan struct{})
	go func() {
		err := run.Wait()
		status := run.ProcessState.Sys().(syscall.WaitStatus)
		signaled := status.Signaled()
		if !signaled {
			log.Fatal(err)
		}
		close(done)
	}()
	go func() {
		<-ctx.Done()
		log.Printf("Stopping gRPC Server. pid: %d\n", run.Process.Pid)
		run.Process.Kill()
	}()
	return done, nil
}

type Options struct {
	Port     string
	BindAddr string
	Output   string
}

const defaultUploadPort = "4772"

func uploadServer(ctx context.Context, opt Options, srvrs *servers) <-chan struct{} {
	if opt.Port == "" {
		opt.Port = defaultUploadPort
	}
	addr := opt.BindAddr + ":" + opt.Port
	r := chi.NewRouter()
	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error Retrieving the File: %v\n", err)
			failed(w, err)
			return
		}

		// shutdown
		srvrs.shutdown()

		err = Unzip(body, opt.Output)
		if err != nil {
			failed(w, err)
			return
		}

		// boot
		err = srvrs.boot(ctx)
		if err != nil {
			failed(w, err)
		}
	})

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

func failed(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

// Unzip will decompress a zip bytes, moving all files and folders
// within the zip bytes (parameter 1) to an output directory (parameter 2).
func Unzip(src []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		log.Fatal(err)
	}

	var filenames []string

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fPath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fPath)
		}

		filenames = append(filenames, fPath)

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

func importSudDirs(names []string) []string {
	dirs := []string{}
	for _, name := range names {
		files, err := ioutil.ReadDir(name)
		if err != nil {
			log.Fatalf("Can't read dir for imports from %s. %v\n", name, err)
		}

		for _, fi := range files {
			switch mode := fi.Mode(); {
			case mode.IsDir():
				dirs = append(dirs, filepath.Join(name, fi.Name()))
			}
		}
	}
	return dirs
}

func expandDirs(names []string) ([]string, []string) {
	paths := []string{}
	imps := []string{}
	for _, name := range names {
		fi, err := os.Stat(name)
		if err != nil {
			log.Println(err)
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			imps = append(imps, name)
			err := filepath.Walk(name,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.Mode().IsDir() || !strings.HasSuffix(info.Name(), ".proto") {
						return nil
					}
					paths = append(paths, path)
					return nil
				})
			if err != nil {
				log.Fatal(err)
			}
		case mode.IsRegular() && strings.HasSuffix(name, ".proto"):
			paths = append(paths, name)
		}
	}
	return paths, imps
}
