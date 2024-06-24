module github.com/avran02/fileshare/gateway

go 1.22.3

require (
	github.com/avran02/fileshare/files v0.0.0-20240615204757-7cac6a6a6456
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/chi/v5 v5.0.12
	github.com/json-iterator/go v1.1.12
	google.golang.org/grpc v1.64.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/avran02/fileshare/files => ../files
