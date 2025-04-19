ENV_FILE = .env

ENV_VARS = \
    POSTGRES_DB=avito \
    POSTGRES_USER=user \
    POSTGRES_PASSWORD=password \
    POSTGRES_HOST=db \
    POSTGRES_PORT=5432 \
	PGBOUNCER_HOST=pgbouncer\
	PGBOUNCER_PORT=6432\
	DB_HOST=pgbouncer\
	DB_PORT=6432\
	DB_USER=user\
	DB_PASSWORD=password\
	DB_NAME=avito\
	SSL_MODE=disable \
	JWTSECRET=dontHackMePls \
	
PROTO_DIR=proto/v1/
OUT_DIR=internal/gen/proto

env:
	$(eval SHELL:=/bin/bash)
	printf "%s\n" $(ENV_VARS) > $(ENV_FILE)
	echo "$(ENV_FILE) file created"

run:
	docker compose up --build

runl:
	go run cmd/pvz/main.go

off:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs

lint:
	golangci-lint run

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

cover-html:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

test:
	go test -v ./...
test-int:
	go test -v -tags=integration ./tests/integration_test
gen:
	go generate ./...
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o internal/gen/oapi/dto.go -generate types -package oapi api/v1/swagger.yaml
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o internal/gen/oapi/server.go -generate fiber-server -package oapi api/v1/swagger.yaml

gen-proto:
	protoc \
      --proto_path=$(PROTO_DIR) \
      --go_out=$(OUT_DIR) --go_opt=paths=source_relative \
      --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
      $(PROTO_DIR)/pvz.proto