build:
	@go build -o ./bin/client ./cmd/client/client.go
	@go build -o ./bin/server ./cmd/server/server.go