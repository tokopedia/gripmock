module github.com/tokopedia/gripmock

go 1.18

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/icrowley/fake v0.0.0-20221112152111-d7b7e2276db2
	github.com/lithammer/fuzzysearch v1.1.1
	github.com/stretchr/testify v1.7.5
	github.com/tokopedia/gripmock/protogen/example v0.0.0
	google.golang.org/grpc v1.47.0
	github.com/tokopedia/gripmock/protogen v0.0.0
)

require (
	github.com/corpix/uarand v0.0.0-20170723150923-031be390f409 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201111145450-ac7456db90a6 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// this is for generated server to be able to run
replace github.com/tokopedia/gripmock/protogen/example v0.0.0 => ./protogen/example

// this is for example client to be able to run
replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
