BIN := ethdont
CONTAINER=$(BIN)
REPO=poupas/$(CONTAINER)
PLATFORMS=linux/amd64
# Uncomment the following line to build for multiple platforms
#PLATFORMS=linux/arm64,linux/amd64

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

.PHONY: docker-prepare
docker-prepare:
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx create --name multiarch --driver docker-container
	docker buildx inspect --builder multiarch --bootstrap

.PHONY: docker-publish
docker-publish: check-env
	docker buildx build --builder multiarch --platform $(PLATFORMS) \
		-t $(REPO):latest -t $(REPO):$(VERSION) \
		--push .

.PHONY: docker-build
docker-build: check-env
	docker buildx build --builder multiarch --platform $(PLATFORMS) \
		-t $(REPO):latest \
		--load \
		.

.PHONY: docker-clean
docker-clean:
	docker buildx --builder multiarch prune || true
	docker buildx rm multiarch || true

.PHONY: check-env
check-env:
ifndef VERSION
	$(error VERSION env variable is undefined)
endif
