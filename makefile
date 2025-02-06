build:
	@./script/build_all.sh

runs:
	@go run ./cmd/server/server.go -c ./internal/config/dev.yaml

runc:
	@go run ./cmd/client/client.go

fmt:
	@go fmt ./...

clean:
	@rm -rf ./bin/*