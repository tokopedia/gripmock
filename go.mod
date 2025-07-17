module github.com/tokopedia/gripmock

go 1.23.0

toolchain go1.23.8

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.4
	github.com/lithammer/fuzzysearch v1.1.5
	github.com/stretchr/testify v1.7.0
	github.com/tokopedia/gripmock/protogen v0.0.0
	golang.org/x/text v0.23.0
	google.golang.org/grpc v1.72.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
