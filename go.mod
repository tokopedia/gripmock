module github.com/tokopedia/gripmock

go 1.21

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.3
	github.com/lithammer/fuzzysearch v1.1.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.3
	github.com/tokopedia/gripmock/protogen/example v0.0.0
	google.golang.org/grpc v1.55.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// this is for generated server to be able to run
replace github.com/tokopedia/gripmock/protogen/example v0.0.0 => ./protogen/example

// this is for example client to be able to run
replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
