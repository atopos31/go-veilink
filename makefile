build:
	@CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_client_$(os)_$(arch)$(suffix) ./cmd/client/client.go
	@CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -o ./bin/veilink_server_$(os)_$(arch)$(suffix) ./cmd/server/server.go
	@echo "Build success see ./bin/veilink_server_$(os)_$(arch)$(suffix)"

runs:
	@go run ./cmd/server/server.go -c ./internal/config/dev.yaml

runc:
	@go run ./cmd/client/client.go

os ?= linux
arch ?= amd64

suffix = $(if $(filter windows,$(os)),.exe,)

fmt:
	@go fmt ./...

clean:
	@rm -rf ./bin/*