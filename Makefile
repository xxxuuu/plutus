export BUILD_AT := $(shell date -u +'%Y%m%d-%H%M%S')

## generate: Generate code
.PHONY: generate
generate:
	@go generate ./...

## image: Build docker image
.PHONY: image
image:
	docker build -t plutus-$(BUILD_AT) .

## test: Run tests
.PHONY: test
test:
	@anvil --fork-url=https://bscrpc.com > /dev/null &
	@go test -v ./...
	@kill $(shell pgrep -f anvil)
