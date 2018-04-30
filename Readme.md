# GripMock
GripMock is a **mock server** for **GRPC** services. It's using `.proto` file to generate implementation of gRPC service for you.
If you already familiar with [Apiary](https://apiary.io) or [WireMock](http://wiremock.org) for mocking API service and looking for similiar thing for GRPC then this is the perfect fit for that.


# How It Works
GripMock will serve *TCP* on port `:4770` for GRPC service and *HTTP* on port `:4771` for Stub service.
When GRPC service receive a request it will find mapped stub in stub service for matching stubs.

Matched stub will be returned to GRPC service as a JSON response then further unmarshal into struct reply.


## Quick Usage
First, prepare your `.proto` file. or you can use `hello.proto` in `example/pb/` folder. Suppose you put it in `/mypath/hello.proto`. We gonna use Docker image for easier example test.
basic syntax to run GripMock is 
`gripmock <protofile>`

- Install [Docker](https://docs.docker.com/install/)
- Run `docker pull ahmadmuzakki/gripmock` to pull the image
- We gonna mount `/mypath/hello.proto` (it must a fullpath) into container. Run `docker run -v /mypath:/proto ahmadmuzakki/gripmock gripmock /proto/hello.proto`
- On separate terminal we gonna add stub into stub service. Run `curl -X POST -d '{"service":"Greeter","method":"SayHello","input":{"equals":{"name":"gripmock"}},"output":{"data":{"message":"Hello GripMock"}}}' localhost:4771/add `
- Now we are ready to test it with our client. you can find client example file under `example/client/`. Execute one of your preferred language. Example for go: `go run example/client/go/main.go`

## Stubbing

Stubbing is the essential mocking of GripMock. It will match and return the expected result into GRPC service. This is where you put all your request expectation and response

Stub Server running on `http://localhost:4771`

- `GET /` Will list all stubs mapping.
- `POST /add` Will add stub with provided stub data
- `POST /find` Find matching stub with provided input. see Input Matching below.

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

## Input Matching
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
    "contains":{
      "name":"^grip.*$"
    }
  }
  .
  .
}
```

