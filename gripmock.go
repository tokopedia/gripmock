package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/quintans/gripmock/servers"
	"github.com/quintans/gripmock/stub"
	"github.com/quintans/gripmock/upload"
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
		log.Println("version:", servers.Version)
		return
	}

	flag.Parse()
	log.Println("Starting GripMock", servers.Version)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("GOPATH is empty")
	}
	if *output == "" {
		*output = os.Getenv("GOPATH") + "/src/grpc"
	}
	upProtoFolder := goPath + "/temp/up-proto"
	// for safety
	*output += "/"
	if _, err := os.Stat(*output); os.IsNotExist(err) {
		os.Mkdir(*output, os.ModePerm)
	}

	importDirs := []string{"/protobuf"}
	impSplit := strings.Split(*imports, ",")
	if len(*imports) > 0 {
		importDirs = append(importDirs, impSplit...)
	}

	// parse proto files
	args := flag.Args()
	protoPaths, imps := servers.ExpandDirs(args)

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
	srvrs := servers.New(
		goPath,
		stub.Options{
			StubPath: *stubPath,
			Port:     *adminport,
			BindAddr: *adminBindAddr,
		},
		servers.ProtocParam{
			ProtoPaths:    protoPaths,
			AdminPort:     *adminport,
			GrpcAddress:   *grpcBindAddr,
			GrpcPort:      *grpcPort,
			Imports:       importDirs,
			ImportSubdirs: *importSubdirs,
		},
		*output,
		upProtoFolder,
	)

	err := srvrs.CleanUploadDir()
	if err != nil {
		log.Fatalf("Unable to clean upload dir: %v", err)
	}

	err = srvrs.Boot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	upDone := upload.RunUploadServer(ctx, upload.Options{
		BindAddr: *upBindAddr,
		Port:     *upPort,
	}, srvrs)

	var term = make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case <-term:
		log.Println("Signaled to shutdown")
		cancel()
		srvrs.Shutdown()
		<-upDone
	}
}
