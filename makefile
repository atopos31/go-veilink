build:
	@go build -o ./bin/main ./cmd/client/client.go
	@go build -o ./bin/server ./cmd/server/server.go