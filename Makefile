.PHONY: run test verify container
run:
	go run ./cmd/worker
test:
	go test -race ./...
verify:
	test -z "$$(gofmt -l .)"
	go test -race ./...
container:
	docker build -t fluxmesh .
