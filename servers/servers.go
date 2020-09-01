package servers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tokopedia/gripmock/stub"
)

//go:generate mockgen -source=servers.go -destination mocks/servers.go -package=mocks Rebooter

const (
	binaryName = "grpcserver"
)

type Rebooter interface {
	Boot(ctx context.Context) error
	Shutdown()
	Reset(reset Reset)
	UploadDir() string
	CleanUploadDir() error
}

type Servers struct {
	goPath     string
	options    stub.Options
	params     ProtocParam
	grpcOutput string
	uploadDir  string
	cancel     context.CancelFunc
	stubDone   <-chan struct{}
	grpcDone   <-chan struct{}
}

type ProtocParam struct {
	ProtoPaths    []string
	AdminPort     string
	GrpcAddress   string
	GrpcPort      string
	Imports       []string
	ImportSubdirs bool
}

type Reset struct {
	ImportSubDirs bool `json:"isd"`
}

func New(goPath string, options stub.Options, params ProtocParam, grpcOutput string, uploadDir string) *Servers {
	return &Servers{
		goPath:     goPath,
		options:    options,
		params:     params,
		grpcOutput: grpcOutput,
		uploadDir:  uploadDir,
	}
}

func (s *Servers) UploadDir() string {
	return s.uploadDir
}

func (s *Servers) Boot(ctx context.Context) error {
	protos, _ := ExpandDirs([]string{s.uploadDir})
	protoPaths := append(s.params.ProtoPaths, protos...)
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
	if s.params.ImportSubdirs {
		importDirs = append(s.params.Imports, ImportSudDirs([]string{s.uploadDir})...)
	} else {
		importDirs = append(s.params.Imports, s.uploadDir)
	}

	pp := ProtocParam{
		Imports:     importDirs,
		AdminPort:   s.params.AdminPort,
		GrpcAddress: s.params.GrpcAddress,
		GrpcPort:    s.params.GrpcPort,
		ProtoPaths:  protoPaths,
	}

	// generate pb.go and grpc server based on proto
	err = generateProtoc(s.goPath, pp, s.grpcOutput)
	if err != nil {
		return err
	}

	// build the server
	err = buildServer(s.goPath, s.grpcOutput)
	if err != nil {
		return err
	}

	// and run
	s.grpcDone, err = runGrpcServer(ctx, s.grpcOutput)
	return err
}

func (s *Servers) CleanUploadDir() error {
	// clear output upload folder
	err := os.RemoveAll(s.uploadDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(s.uploadDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Servers) Shutdown() {
	if s.cancel != nil {
		s.cancel()
		<-s.stubDone
		<-s.grpcDone
	}
}

func (s *Servers) Reset(reset Reset) {
	s.params.ImportSubdirs = reset.ImportSubDirs
}

func generateProtoc(goPath string, other ProtocParam, output string) error {
	src := goPath + "/src"
	args := []string{}
	// include well-known-types
	for _, i := range other.Imports {
		args = append(args, "-I", i)
		log.Println("Importing", i)
	}
	args = append(args, "--go_out=plugins=grpc:"+src)
	args = append(args, fmt.Sprintf("--gripmock_out=admin-port=%s,grpc-address=%s,grpc-port=%s:%s",
		other.AdminPort, other.GrpcAddress, other.GrpcPort, output))
	args = append(args, other.ProtoPaths...)
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

func buildServer(goPath string, output string) error {
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

func ImportSudDirs(names []string) []string {
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

func ExpandDirs(names []string) ([]string, []string) {
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
