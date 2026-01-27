.PHONY: run test mock

run:
	go run cmd/server/main.go

test:
	go test ./...

mock:
	# mockgen command
