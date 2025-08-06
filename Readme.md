# GripMock
GripMock is a **mock server** for **gRPC** services. It uses a `.proto` file to generate implementation of gRPC service for you.
You can use gripmock for setting up end-to-end testing or as a dummy server in a software development phase.
The server implementation is in GoLang but the client can be in any programming language that supports gRPC.

---

### Note from the author
Hi all, Jackie here. First of all, I would like to thank you for all those who have contributed to this project. This project has been and will be maintained in my free time, and I maintain it basically by myself, so I apologize for any extremely delayed responses from my side. 

Regarding contribution:
- Please raise a PR with a detailed description and state a clear motivation. 
- The code implementation should meet the standards that have been set in this repo. 
- Only well-demanded features will be accepted, so if you think a feature is uniquely suited to your use case, you can keep it to yourself.
- Tests are mandatory, both unit tests and integration tests whenever applicable. 

Regarding roadmap:
- I would like to keep gripmock as simple as possible. I want this tool to be as reliable as possible and easy to maintain.
- My priority is to keep up with the latest updates from the frameworks (go, grpc, protoc, etc.) so that this tool can be used with the latest proto.
- Once in a while, I'll check the issues and PR list to see what features are worth adding to gripmock. 

---

## Quick Usage
First, prepare your `.proto` file. Or you can use `hello.proto` in `example/simple/` folder. Suppose you put it in `/mypath/hello.proto`. We will use Docker image for easier example test.

- Install [Docker](https://docs.docker.com/install/)
- Run `docker pull tkpd/gripmock` to pull the image
- We will mount `/mypath/hello.proto` (it must be a fullpath) into a container and also we expose ports needed. Run `docker run -p 4770:4770 -p 4771:4771 -v /mypath:/proto tkpd/gripmock /proto/hello.proto`
- On a separate terminal, we will add a stub into the stub service. Run `curl -X POST -d '{"service":"Gripmock","method":"SayHello","input":{"equals":{"name":"gripmock"}},"output":{"data":{"message":"Hello GripMock"}}}' localhost:4771/add `
- Now we are ready to test it with our client. You can find a client example file under `example/simple/client/`. Execute the example in your preferred language. Example for Go: `go run example/simple/client/*.go`

Check [`example`](https://github.com/tokopedia/gripmock/tree/master/example) folder for various use cases of gripmock.

---

## How It Works
![Running Gripmock](/assets/images/gripmock_readme-running%20system.png)

From client perspective, GripMock has 2 main components:
1. GRPC server that serves on `tcp://localhost:4770`. Its main job is to serve incoming RPC calls from client and then parse the input so that it can be posted to Stub service to find the perfect stub match.
2. Stub server that serves on `http://localhost:4771`. Its main job is to store all the stub mappings. We can add a new stub or list existing stubs using http request.

Matched stub will be returned to GRPC service then further parse it to respond to the RPC call.


From technical perspective, GripMock consists of 2 binaries. 
The first binary is the gripmock itself, which will generate the gRPC server using the plugin installed in the system (see [Dockerfile](Dockerfile)). 
When the server successfully generated, it will be invoked in parallel with stub server which ends up opening 2 ports for client to use.

The second binary is the protoc plugin which is located in the folder [protoc-gen-gripmock](/protoc-gen-gripmock). This plugin is the one who translates protobuf declaration into a gRPC server in Go programming language. 

![Inside GripMock](/assets/images/gripmock_readme-inside.png)

---

## Stubbing

Stubbing is the essential mocking feature of GripMock. It will match and return the expected result into GRPC service. This is where you put all your request expectations and responses.

### Dynamic stubbing
You could add stubbing on the fly with a simple REST API. HTTP stub server is running on port `:4771`

- `GET /` Will list all stubs mapping.
- `POST /add` Will add stub with provided stub data
- `POST /find` Find matching stub with provided input. see [Input Matching](#input_matching) below.
- `GET /clear` Clear stub mappings.
- `POST /reset` Reset stub mappings by clearing all stubs and reloading them from the configured stub file path (if provided).
- `GET /requests` List all recorded requests that have been made to the stub server.

Stub Format is JSON text format. It has a skeleton as follows:
```
{
  "service":"<servicename>", // name of service defined in proto
  "method":"<methodname>", // name of method that we want to mock
  "input":{ // input matching rule. see Input Matching Rule section below
    // put rule here
  },
  "headers":{ // Optional header matching rule. See "Headers Matching Rule" section below.
    // put rule here
  },
  "output":{ // output json if input were matched
    "data":{
      // put result fields here
    },
    "headers": {
      // put result headers here
    },
    "error":"<error message>" // Optional. if you want to return error instead.
    "code":"<response code>" // Optional. Grpc response code. if code !=0  return error instead.
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
    // input that is supposed to match with stored stubs
  }
}
```
So if you do a `curl -X POST -d '{"service":"Greeter","method":"SayHello","data":{"name":"gripmock"}}' localhost:4771/find` stub service will find a match from listed stubs.

### Input Matching Rule
Input matching has 4 rules to match an input: **equals**, **equals_unordered**, **contains** and **regex**
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
      "numbers": [4, 8, 15, 16, 23, 42],
      "null": null
    }
  }
  .
  .
}
```

**equals_unordered** will match the exact field name and value of input into expected stub, except lists (which are compared as sets). example stub JSON:


```
{
  .
  .
  "input":{
    "equals_unordered":{
      "name":"gripmock",
      "greetings": {
            "english": "Hello World!",
            "indonesian": "Halo Dunia!",
            "turkish": "Merhaba Dünya!"
      },
      "ok": true,
      "numbers": [4, 8, 15, 16, 23, 42],
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

### Headers Matching Rule

Input headers matching has 4 rules to match input headers: `equals`, `equals_unordered`, `contains`, and `matches`.
<br>
Headers are map of strings and the same matching rules apply for headers as for payload data.
<br>
**Important Note:** Only one rule type is applied at a time. If multiple rule types are specified, the first matching rule will be used.
<br>
Headers can be specified in the `headers` field of the root object:
<br>
```json
{
  "service": "YourService",
  "method": "YourMethod",
  "input": {
    "equals": {
      "field1": "value1"
    },
  },
  "headers": {
    "equals": {
      "Content-Type": "application/json"
    }
    // Only one rule type is applied. If you specify multiple rules,
    // the first matching rule will be used.
  },
  "output": {
    "data": {
      "result": "success"
    }
  }
}
```
