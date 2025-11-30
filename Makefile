.PHONY: run test lint format frontend-install

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	go vet ./...
	npm --prefix web run lint

format:
	gofmt -w ./cmd ./internal
	npm --prefix web run format

frontend-install:
	npm --prefix web install
