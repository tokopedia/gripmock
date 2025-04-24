module github.com/tokopedia/gripmock

go 1.21

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.3
	github.com/lithammer/fuzzysearch v1.1.5
	github.com/stretchr/testify v1.7.0
	github.com/tokopedia/gripmock/protogen v0.0.0
	golang.org/x/text v0.14.0
	google.golang.org/grpc v1.62.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
