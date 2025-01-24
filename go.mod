module github.com/tokopedia/gripmock

go 1.23.5

require (
	github.com/go-chi/chi/v5 v5.2.0
	github.com/lithammer/fuzzysearch v1.1.8
	github.com/stretchr/testify v1.7.0
	github.com/tokopedia/gripmock/protogen/example v0.0.0
	golang.org/x/text v0.21.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/genproto v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

// this is for generated server to be able to run
replace github.com/tokopedia/gripmock/protogen/example v0.0.0 => ./protogen/example

// this is for example client to be able to run
replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
