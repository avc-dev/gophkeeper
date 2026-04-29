VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

MIGRATION_DSN ?= postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable
TEST_MIGRATION_DSN ?= postgres://gophkeeper:gophkeeper@localhost:5433/gophkeeper_test?sslmode=disable

# Пакеты с юнит-тестами (не включает cmd/*, proto/*, server/storage/* — они покрываются интеграционными тестами).
UNIT_PKGS = \
	./internal/crypto/... \
	./internal/client/service/... \
	./internal/client/storage/... \
	./internal/server/service/... \
	./internal/server/handler/...

COVERAGE_THRESHOLD = 70

.PHONY: build build-all test-unit test-integration lint proto migrate-up migrate-down migrate-test-up

## Сборка сервера и клиента для текущей платформы
build:
	go build $(LDFLAGS) -o bin/server ./cmd/server
	go build $(LDFLAGS) -o bin/client ./cmd/client

## Сборка для всех платформ
build-all:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-server-linux-amd64   ./cmd/server
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-linux-amd64          ./cmd/client
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-server-darwin-amd64  ./cmd/server
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-darwin-amd64         ./cmd/client
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/gophkeeper-server-darwin-arm64  ./cmd/server
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/gophkeeper-darwin-arm64         ./cmd/client
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-server-windows.exe   ./cmd/server
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/gophkeeper-windows.exe          ./cmd/client

## Юнит-тесты с покрытием (только пакеты с юнит-тестами).
## cmd/*, proto/*, server/storage/* не учитываются: они требуют реальной БД (интеграционные тесты).
test-unit:
	go test $(UNIT_PKGS) -count=1 -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep total
	@go tool cover -func=coverage.out | awk '/total/ {gsub("%",""); if ($$3+0 < $(COVERAGE_THRESHOLD)) \
		{print "ERROR: coverage " $$3 "% is below minimum " $(COVERAGE_THRESHOLD) "%"; exit 1}}'

## Интеграционные тесты (требует запущенного docker-compose.test.yml)
test-integration:
	docker compose -f docker-compose.test.yml up -d --wait
	migrate -path migrations -database "$(TEST_MIGRATION_DSN)" up
	go test ./tests/integration/... -tags=integration -count=1 -timeout=120s
	docker compose -f docker-compose.test.yml down

## Линтер
lint:
	golangci-lint run ./...

## Генерация кода из proto файлов
proto:
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

## Накатить миграции (разработка)
migrate-up:
	migrate -path migrations -database "$(MIGRATION_DSN)" up

## Откатить последнюю миграцию (разработка)
migrate-down:
	migrate -path migrations -database "$(MIGRATION_DSN)" down 1

## Накатить миграции (тесты)
migrate-test-up:
	migrate -path migrations -database "$(TEST_MIGRATION_DSN)" up
