# GripMock
GripMock is a **mock server** for **GRPC** services. It's using `.proto` file to generate implementation of gRPC service for you.
If you already familiar with [Apiary](https://apiary.io) or [WireMock](http://wiremock.org) for mocking API service and looking for similiar thing for GRPC then this is the perfect fit for that.


## How It Works
GripMock has 2 main components:
1. GRPC server that serving on `tcp://localhost:4770`. It's main job is to serve incoming rpc call from client then parse the input so that can be posted to Stub service to find the perfect stub match.
2. Stub server that serving on `http://localhost:4771`. It's main job is to store all the stub mapping. We can add a new stub or list existing stub using http request.

Matched stub will be returned to GRPC service then further parse it to response the rpc call.

## Quick Usage
First, prepare your `.proto` file. or you can use `hello.proto` in `example/pb/` folder. Suppose you put it in `/mypath/hello.proto`. We gonna use Docker image for easier example test.
basic syntax to run GripMock is 
`gripmock <protofile>`

- Install [Docker](https://docs.docker.com/install/)
- Run `docker pull quintans/gripmock` to pull the image
- We gonna mount `/mypath/hello.proto` (it must a fullpath) into container and also we expose ports needed. Run `docker run -p 4770:4770 -p 4771:4771 -v /mypath:/proto quintans/gripmock /proto/hello.proto`
- On separate terminal we gonna add stub into stub service. Run `curl -X POST -d '{"service":"Greeter","method":"SayHello","input":{"equals":{"name":"gripmock"}},"output":{"data":{"message":"Hello GripMock"}}}' localhost:4771/add `
- Now we are ready to test it with our client. you can find client example file under `example/client/`. Execute one of your preferred language. Example for go: `go run example/client/go/*.go`

Check [`example`](https://github.com/quintans/gripmock/tree/master/example) folder for various usecase of gripmock.

## Stubbing

Stubbing is the essential mocking of GripMock. It will match and return the expected result into GRPC service. This is where you put all your request expectation and response

### Dynamic stubbing
You could add stubbing on the fly with simple REST. HTTP stub server running on port `:4771`

- `GET /` Will list all stubs mapping.
- `POST /add` Will add stub with provided stub data
- `POST /find` Find matching stub with provided input. see [Input Matching](#input_matching) below.
- `GET /clear` Clear stub mappings.

Stub Format is JSON text format. It has skeleton like below:
```
{
  "service":"<servicename>", // name of service defined in proto
  "method":"<methodname>", // name of method that we want to mock
  "input":{ // input matching rule. see Input Matching Rule section below
    // put rule here
  },
  "output":{ // output json if input were matched
    "data":{
      // put result fields here
    },
    "error":"<error message>" // Optional. if you want to return error instead.
  }
}
```

For our `hello` service example we put stub with below text:
```
  {
    "service":"Greeter",
    "method":"SayHello",
    "input":{
      "equals":{
        "name":"gripmock"
      }
    },
    "output":{
      "data":{
        "message":"Hello GripMock"
      }
    }
  }
```

### Static stubbing
You could initialize gripmock with stub json files and provide the path using `--stub` argument. For example you may 
mount  your stub file in `/mystubs` folder then mount it to docker like
 
 `docker run -p 4770:4770 -p 4771:4771 -v /mypath:/proto -v /mystubs:/stub quintans/gripmock --stub=/stub /proto/hello.proto`
 
Please note that Gripmock still serve http stubbing to modify stored stubs on the fly.
 
## <a name="input_matching"></a>Input Matching
Stub will responding the expected response if only requested with matching rule of input. Stub service will serve `/find` endpoint with format:
```
{
  "service":"<service name>",
  "method":"<method name>",
  "data":{
    // input that suppose to match with stored stubs
  }
}
```
So if you do `curl -X POST -d '{"service":"Greeter","method":"SayHello","data":{"name":"gripmock"}}' localhost:4771/find` stub service will find a match from listed stubs stored there.

### Input Matching Rule
Input matching has 3 rules to match an input. which is **equals**,**contains** and **regex**
<br>
**equals** will match the exact field name and value of input into expected stub. example stub JSON:
```
{
  .
  .
  "input":{
    "equals":{
      "name":"gripmock"
    }
  }
  .
  .
}
```

**contains** will match input that has the value declared expected fields. example stub JSON:
```
{
  .
  .
  "input":{
    "contains":{
      "field2":"hello"
    }
  }
  .
  .
}
```

**matches** using regex for matching fields expectation. example:

```
{
  .
  .
  "input":{
    "matches":{
      "name":"^grip.*$"
    }
  }
  .
  .
}
```

## After Fork
The above is still valid.

The reason I created a fork was because I did a lot of changes and ended up drifting a bit from the original project.

## New features

This fork adds the following features:
* Uploading proto files
* When uploading protos, immediate sub directories can represent different imports, if tthe flag `--isd` is set
* Packages at the same Level - proto files defined at their own folders and none of them is at the min package (top level)

Changes:
* modules
* generating proto into GOPATH
* Argument list now considers directories. These directories are added to the import list and all the proto files inside are also added to the list of compiled protos.

Breaking changes:
* `GOPATH` needs to be defined


### Packages at the same Level
Now we can address the scenario where we have proto files packages where none is at the main package.
Imagine that we have two projects. One is `foo` and the other is `bar`. `foo` has a dependency on `bar`.
We should do the following.
Create a folder that will have all proto files (eg: proto) and then underneath the folders for the proto files for `foo` and another for `bar`.

We will end up with:

```
proto
├── bar
│   └── bar.proto
└── foo
    ├── foo.proto
    └── hello.proto
```

Unfortunately just copying doesn't work for all projects, since some projects have their own particularities and a generic tool would have a hard time trying to cope with all of them.
I am thinking about when some projects define the packages in the `protoc` command like `--go_out=Mbar/bar.proto=this/is/a/package:.`.
What we have to do is the opposite. Change the proto files used for the mocks so that all have the option `go_package` and that the imports reflect the current structure.

If sub dirs import flag is set `-isb`, all immediate sub dirs from the uploaded protos.
Consider a zip file with the following tree dir.

```
proto
├── prj-bar
│   └── bar
│       ├── bar.pb.go
│       └── bar.proto
└── prj-foo
    └── foo
        ├── foo.pb.go
        ├── foo.proto
        ├── hello.pb.go
        └── hello.prot
``` 

`prj-foo` and `prj-bar` inside `/proto`, will be imported by the `protoc` with the compile option `-I` allowing us to have different packages in different sub directories.

Please make you proto well behaved:
* all proto files have the `option go_package`.
  * eg: `option go_package = "github.com/quintans/foo";`
* all the imports must be relative to the current file structure
  * eg: `import "github.com/quintans/bar/bar.proto";` ==> `import "bar/bar.proto";`
* do not use `.` in the package name.
  * eg: `svc.foo` => `foo`. It is best to make it consistent with the folder name.

Finally start gripmock specifying the top level proto folder
```sh
gripmock -o ./grpc ./proto
```

or without an initial proto folder 

```sh
gripmock -o ./grpc
```

Docker
```sh
docker run -p 4770:4770 -p 4771:4771 -p 4772:4772 -v /mypath:/proto -v /mystubs:/stub quintans/gripmock --stub=/stub /proto
```

or without an initial proto folder or stubs

```sh
docker run -p 4770:4770 -p 4771:4771 -p 4772:4772 quintans/gripmock
```

### Uploading proto files
Proto files can be uploaded by zipping the proto folder and upload it to `http://localhost:4772/upload`. There is also an utility tool at the `tool` package. 

This will add up to the existing proto files.

## Troubleshooting

When an error occurs it have the detail of what went wrong.
If that is not enough you can browse the image file system through `http://localhost:4772/dir/`

`/go/src/grpc` is the default location of the generated server code `server.go`.

If you still can't figure out what is going on you can always run this project and try debug the problem.

```sh
svc-gripmock --imports=path/to/protos/folder1,path/to/protos/folder2 path/to/protos
```

then you can look in `$GOPATH/src/grpc` for the generated server
