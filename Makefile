run:
	@go run ./cmd

test:
	@go test .

format:
	@gofmt -w -l .
