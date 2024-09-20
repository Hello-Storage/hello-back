# Copyright © 2023 Hello Storage Inc. All rights reserved.
#
export GO111MODULE=on
export DOCKER_BUILDKIT=1


# 
APP_NAME=hello
DOCKER_COMPOSE=docker compose
GOTEST=go test

run: 
	go run cmd/main.go
dev:
	air -c .air.toml
# 
develop:
	$(DOCKER_COMPOSE) -f compose.dev.yml up -d --build --wait
stop-develop:
	$(DOCKER_COMPOSE) -f compose.dev.yml stop
down-develop:
	$(DOCKER_COMPOSE) -f compose.dev.yml down --volumes
production:
	$(DOCKER_COMPOSE) -f compose.prod.yml up -d --build
stop-production:
	$(DOCKER_COMPOSE) -f compose.prod.yml stop
build:
	scripts/docker/build.sh develop
buildx:
	scripts/docker/buildx.sh develop
logs:
	$(DOCKER_COMPOSE) logs -f
build-go:
	rm -f build/$(APP_NAME)
	scripts/build.sh debug $(APP_NAME)
test:
	$(info Running all Go tests...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
fmt:
	go fmt ./pkg/... ./internal/... ./cmd/...
	gofmt -w -s pkg internal cmd
	goimports -w pkg internal cmd
install:
	go get .
tidy:
	go mod tidy
clean:
	docker image prune

# TO-DO drone configure
# drone:


.PHONY: all test clean