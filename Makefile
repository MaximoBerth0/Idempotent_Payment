build:
	go build ./cmd

run:
	go run ./cmd

test:
	go test ./...

lint:
	golangci-lint run