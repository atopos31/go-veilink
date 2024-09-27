build:
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_client_$(os)_$(arch)$(suffix) ./cmd/client/client.go
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_server_$(os)_$(arch)$(suffix) ./cmd/server/server.go
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_socks5_$(os)_$(arch)$(suffix) ./cmd/client/socks5.go

runs:
	@go run ./cmd/server/server.go -c ./internal/config/server.toml

runc:
	@go run ./cmd/client/client.go

runsc:
	@go run ./cmd/socks5/socks5.go -ip=127.0.0.1 -port=1080 -level=debug 

os ?= linux
arch ?= amd64

suffix = $(if $(filter windows,$(os)),.exe,)

fmt:
	@go fmt ./...

clean:
	@rm -rf ./bin/*