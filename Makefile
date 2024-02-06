build:
	@go build -o ./bin/server ./cmd

run:
	@go run ./cmd

test:
	@go test .

format:
	@gofmt -w -l .
