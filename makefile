build:
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_client_$(os)_$(arch) ./cmd/client/client.go
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_server_$(os)_$(arch) ./cmd/server/server.go

os ?= linux
arch ?= amd64