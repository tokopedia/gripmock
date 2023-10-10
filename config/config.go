package config

import (
	"flag"
	"log"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	GrpcPort    string   `mapstructure:"GRPC_PORT"`
	GrpcListen  string   `mapstructure:"GRPC_LISTEN"`
	AdminPort   string   `mapstructure:"ADMIN_PORT"`
	AdminListen string   `mapstructure:"ADMIN_LISTEN"`
	StubsDir    string   `mapstructure:"STUBS_DIR"`
	ProtoDirs   []string `mapstructure:"PROTO_DIRS"`
	WktProto    string   `mapstructure:"WKT_PROTO"`
}

func LoadEnv() (config *Config) {
	viper.AutomaticEnv()
	viper.SetDefault("GRPC_PORT", "4770")
	viper.SetDefault("GRPC_LISTEN", "")
	viper.SetDefault("ADMIN_PORT", "4771")

	viper.SetDefault("STUBS", "")
	viper.SetDefault("ADMIN_LISTEN", "localhost")
	viper.SetDefault("WKT_PROTO", "/protobuf")

	if err := viper.Unmarshal(&config, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToSliceHookFunc(",")))); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	parseCmdArgs(config)
	return
}

func parseCmdArgs(config *Config) {

	grpcPort := flag.String("grpc-port", "", "Port of gRPC tcp server")
	grpcBindAddr := flag.String("grpc-listen", "", "Adress the gRPC server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	adminport := flag.String("admin-port", "", "Port of stub admin server")
	adminBindAddr := flag.String("admin-listen", "", "Adress the admin server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	stubPath := flag.String("stubs-dir", "", "Path where the stub files are (Optional)")
	protos := flag.String("proto-dirs", "", "comma separated imports path")
	wktProto := flag.String("wkt-proto", "", "Path to the well known protos. Default path /protobuf is where gripmock Dockerfile install WKT protos")

	flag.Parse()

	if *grpcPort != "" {
		config.GrpcPort = *grpcPort
	}
	if *grpcBindAddr != "" {
		config.GrpcListen = *grpcBindAddr
	}
	if *adminport != "" {
		config.AdminPort = *adminport
	}
	if *adminBindAddr != "" {
		config.AdminListen = *adminBindAddr
	}
	if *stubPath != "" {
		config.StubsDir = *stubPath
	}
	if *protos != "" {
		config.ProtoDirs = strings.Split(*protos, ",")
	}
	if *wktProto != "" {
		config.WktProto = *wktProto
	}
}
