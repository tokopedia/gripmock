package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tokopedia/gripmock/config"
	"github.com/tokopedia/gripmock/stub"
)

const output = "grpc-mock-server/"
const protogen = output + "protogen/"

func main() {
	config := config.LoadEnv()

	if config == nil {
		log.Fatal("Config is nil")
	}

	fmt.Println("Starting GripMock")

	// run admin stub server
	stub.RunStubServer(stub.Options{
		StubPath: config.StubsDir,
		Port:     config.AdminPort,
		BindAddr: config.AdminListen,
	})

	// generate pb.go and grpc server based on proto
	generateProtoc(*config)

	// build the server
	//buildServer(output)

	// and run
	run, runerr := runGrpcServer(output)

	var term = make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT)
	select {
	case err := <-runerr:
		log.Fatal(err)
	case <-term:
		fmt.Println("Stopping gRPC Server")
		run.Process.Kill()
	}
}

func generateProtoc(config config.Config) {
	protoFiles := prepareProtos(config.ProtoDirs)

	// estimate args length to prevent expand
	args := make([]string, 0, len(config.ProtoDirs)+len(protoFiles)+2)
	protogenDir, err := filepath.Abs(protogen)
	if err != nil {
		log.Fatal("Fail on get abs path ", err)
	}
	args = append(args, "-I", protogenDir)
	if _, err := os.Stat(config.WktProto); err == nil {
		args = append(args, "-I", config.WktProto)
	}
	args = append(args, protoFiles...)
	args = append(args, "--go_out=plugins=grpc:"+output)
	args = append(args, fmt.Sprintf("--gripmock_out=module-name=%s,admin-port=%s,grpc-address=%s,grpc-port=%s:%s",
		"grpcmock",
		config.AdminPort, config.GrpcListen, config.GrpcPort, output))
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	err = protoc.Run()
	if err != nil {
		log.Fatal("Fail on protoc ", err)
	}

}

// append gopackage in proto files if doesn't have any
func prepareProtos(protoDirs []string) (result []string) {
	for _, proto := range protoDirs {
		filepath.WalkDir(proto, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				relativePath, err := filepath.Rel(proto, path)
				if err != nil {
					log.Fatalf("Could not get relative path of %s, error: %v", path, err)
					return err
				}
				dir := filepath.Dir(relativePath)
				dir = strings.TrimLeft(dir, "/")
				fmt.Println(dir)

				// get string from right until the first /
				// example value: hello.proto
				file := filepath.Base(path)
				fmt.Println("Proto file:", file)

				newdir := filepath.Join(protogen, dir)
				if err := os.MkdirAll(newdir, 0750); err != nil {
					log.Fatalf("Could not create dir %s, error: %v", newdir, err)
					return err
				}
				newfile := filepath.Join(newdir, file)

				// copy to protogen directory
				if err := copyFile(path, newfile); err != nil {
					fmt.Println(err)
					return err
				}

				// Force remove any declaration of go_package
				// then replace it with our own declaration below
				removeGoPackageDeclaration(newfile)

				// get the line number of "syntax" declaration
				syntaxLineNum := getSyntaxLineNum(newfile)

				if syntaxLineNum != -1 {

					goPackageString := fmt.Sprintf("option go_package = \"protogen/%s\";", dir)

					// append our own go_package declaration just below "syntax" declaration
					appendGoPackageDeclaration(newfile, syntaxLineNum, goPackageString)
					absolutePath, err := filepath.Abs(newfile)
					if err != nil {
						log.Fatalf("Could not get absolute path of %s, error: %v", newfile, err)
					}
					result = append(result, absolutePath)
				}

			}

			return nil
		})
	}
	return
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func removeGoPackageDeclaration(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	newData := []string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "option go_package") {
			newData = append(newData, line)
		}
	}

	err = os.WriteFile(filename, []byte(strings.Join(newData, "\n")), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func getSyntaxLineNum(filename string) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "syntax") {
			return i + 1
		}
	}

	return -1
}

func appendGoPackageDeclaration(filename string, lineNum int, goPackageString string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	lines := strings.Split(string(data), "\n")
	lines = append(lines[:lineNum], append([]string{goPackageString}, lines[lineNum:]...)...)
	output := strings.Join(lines, "\n")

	err = os.WriteFile(filename, []byte(output), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func runGrpcServer(output string) (*exec.Cmd, <-chan error) {
	var err error
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(output)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Chdir(currentDir)
	goTidy := exec.Command("go", "mod", "tidy")
	goTidy.Stdout = os.Stdout
	goTidy.Stderr = os.Stderr
	err = goTidy.Run()
	if err != nil {
		log.Fatal(err)
	}

	run := exec.Command("go", "run", "server.go")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	err = run.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("grpc server pid: %d\n", run.Process.Pid)
	runerr := make(chan error)
	go func() {
		runerr <- run.Wait()
	}()
	return run, runerr
}
