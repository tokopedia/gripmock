GOLANGCI_LING_IMAGE="golangci/golangci-lint:v1.54.0-alpine"

.PHONY: *

version=latest

build:
	docker buildx build --load -t "bavix/gripmock:${version}" --platform linux/amd64,linux/arm64 .

test:
	go test -tags mock -race -cover ./...

lint:
	docker run --rm -v ./:/app -w /app $(GOLANGCI_LING_IMAGE) golangci-lint run --color always ${args}

lint-fix:
	make lint args=--fix
