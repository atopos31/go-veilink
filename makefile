build:
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_client_$(os)_$(arch)$(suffix) ./cmd/client/client.go
	@GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_server_$(os)_$(arch)$(suffix) ./cmd/server/server.go

os ?= linux
arch ?= amd64

suffix = $(if $(filter windows,$(os)),.exe,)

fmt:
	@go fmt ./...

clean:
	@rm -rf ./bin/*