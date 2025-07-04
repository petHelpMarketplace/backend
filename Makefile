include .env

## COLORS
COLOR_RESET     = \033[0m
COLOR_INFO      = \033[32m
COLOR_COMMENT   = \033[33m
COLOR_ERROR     = \033[0;31m
COLOR_COM       = \033[0;34m
COLOR_OBJ       = \033[0;36m


LOCAL_BIN:=$(CURDIR)/bin

build:
	go build -o bin/main ./cmd/.

dev: ## Start project in the dev mode
	@docker compose up -d
	@go run ./cmd/pethelp .

swag_gen:
	swag init --dir cmd,internal --generalInfo api/main.go --parseInternal 

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

install-goose:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.14.0

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml

local-migration-status:
	${LOCAL_BIN}/goose -dir ${MIGRATION_DIR} postgres "$(PG_DSN)" status

local-migration-up:
	${LOCAL_BIN}/goose -dir ${MIGRATION_DIR} postgres "$(PG_DSN)" up

local-migration-down:
	${LOCAL_BIN}/goose -dir ${MIGRATION_DIR} postgres "$(PG_DSN)" down