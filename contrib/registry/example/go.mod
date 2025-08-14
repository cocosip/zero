module github.com/cocosip/zero/contrib/registry/example

go 1.21

require (
	github.com/cocosip/zero/contrib/registry v0.0.0
	github.com/go-kratos/kratos/v2 v2.8.4
	google.golang.org/protobuf v1.34.2
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/cocosip/zero/contrib/registry => ../

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/cocosip/zero/contrib/registry/local v0.0.0-20250814053226-3cfb19d8db5a // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240617180043-68d350f18fd4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240617180043-68d350f18fd4 // indirect
	google.golang.org/grpc v1.64.1 // indirect
)
