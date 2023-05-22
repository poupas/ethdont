BIN := ethdont
CONTAINER=$(BIN)

.DEFAULT_GOAL := build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: fmt
	@command -v docker >/dev/null || { echo "You need Docker installed to run the linter" && exit 1; }
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint golangci-lint run -v

.PHONY: test
test:
	go test ./services

.PHONY: build
build:
	go build -o $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)

.PHONY: docker
docker:
	docker build -t $(CONTAINER):latest .

.PHONY: check-env
check-env:
ifndef VERSION
	$(error VERSION env variable is undefined)
endif
