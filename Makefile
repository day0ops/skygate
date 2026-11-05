.PHONY: proto build test docker helm-lint run ui-build ui-test

proto:
	buf dep update
	buf generate

build:
	go build ./...

test:
	go test ./...

ui-build:
	cd web && bun install && bun run build

ui-test:
	cd web && bun install && bun run test

docker:
	docker build -t skygate:dev .

helm-lint:
	helm lint charts/skygate

run:
	go run ./cmd/skygate
