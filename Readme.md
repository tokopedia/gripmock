# GripMock
GripMock is a **mock server** for **GRPC** services. It's using a `.proto` file to generate implementation of gRPC service for you.
If you are already familiar with [Apiary](https://apiary.io) or [WireMock](http://wiremock.org) for mocking API service and looking for similiar thing for GRPC then this is the perfect fit for that.


## How It Works
GripMock has 2 main components:
1. GRPC server that serves on `tcp://localhost:4770`. Its main job is to serve incoming rpc call from client and then parse the input so that it can be posted to Stub service to find the perfect stub match.
2. Stub server that serves on `http://localhost:4771`. Its main job is to store all the stub mapping. We can add a new stub or list existing stub using http request.

Matched stub will be returned to GRPC service then further parse it to response the rpc call.

## Quick Usage
First, prepare your `.proto` file. Or you can use `hello.proto` in `example/pb/` folder. Suppose you put it in `/mypath/hello.proto`. We are gonna use Docker image for easier example test.
basic syntax to run GripMock is 
`gripmock <protofile>`

- Install [Docker](https://docs.docker.com/install/)
- Run `docker pull tkpd/gripmock` to pull the image
- We are gonna mount `/mypath/hello.proto` (it must be a fullpath) into a container and also we expose ports needed. Run `docker run -p 4770:4770 -p 4771:4771 -v /mypath:/proto tkpd/gripmock /proto/hello.proto`
- On a separate terminal we are gonna add a stub into the stub service. Run `curl -X POST -d '{"service":"Greeter","method":"SayHello","input":{"equals":{"name":"gripmock"}},"output":{"data":{"message":"Hello GripMock"}}}' localhost:4771/add `
- Now we are ready to test it with our client. You can find a client example file under `example/simple/client/`. Execute one of your preferred language. Example for go: `go run example/simple/client/go/*.go`

Check [`example`](https://github.com/tokopedia/gripmock/tree/master/example) folder for various usecase of gripmock.

## Stubbing

Stubbing is the essential mocking of GripMock. It will match and return the expected result into GRPC service. This is where you put all your request expectation and response

### Dynamic stubbing
You could add stubbing on the fly with a simple REST API. HTTP stub server is running on port `:4771`

- `GET /` Will list all stubs mapping.
- `POST /add` Will add stub with provided stub data
- `POST /find` Find matching stub with provided input. see [Input Matching](#input_matching) below.
- `GET /clear` Clear stub mappings.

Stub Format is JSON text format. It has a skeleton as follows:
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

For our `hello` service example we put a stub with the text below:
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
mount your stub file in `/mystubs` folder then mount it to docker like
 
 `docker run -p 4770:4770 -p 4771:4771 -v /mypath:/proto -v /mystubs:/stub tkpd/gripmock --stub=/stub /proto/hello.proto`
 
Please note that Gripmock still serves http stubbing to modify stored stubs on the fly.
 
## <a name="input_matching"></a>Input Matching
Stub will respond with the expected response only if the request matches any rule. Stub service will serve `/find` endpoint with format:
```
{
  "service":"<service name>",
  "method":"<method name>",
  "data":{
    // input that suppose to match with stored stubs
  }
}
```
So if you do a `curl -X POST -d '{"service":"Greeter","method":"SayHello","data":{"name":"gripmock"}}' localhost:4771/find` stub service will find a match from listed stubs stored there.

### Input Matching Rule
Input matching has 3 rules to match an input: **equals**,**contains** and **regex**
<br>
Nested fields are allowed for input matching too for all JSON data types. (`string`, `bool`, `array`, etc.)
<br>
**Gripmock** recursively goes over the fields and tries to match with given input.
<br>
**equals** will match the exact field name and value of input into expected stub. example stub JSON:
```
{
  .
  .
  "input":{
    "equals":{
      "name":"gripmock",
      "greetings": {
            "english": "Hello World!",
            "indonesian": "Halo Dunia!",
            "turkish": "Merhaba Dünya!"
      },
      "ok": true,
      "numbers": [4, 8, 15, 16, 23, 42]
      "null": null
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
      "field2":"hello",
      "field4":{
        "field5": "value5"
      } 
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
      "name":"^grip.*$",
      "cities": ["Jakarta", "Istanbul", ".*grad$"]
    }
  }
  .
  .
}
```

