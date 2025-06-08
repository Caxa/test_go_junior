include .env
export

.PHONY: test test-cover lint run build clean \
        migrate-up migrate-down migrate-status migrate-force-reset

APP_NAME=go-people-api
TEST_COVERAGE_FILE=coverage.out
TEST_COVERAGE_HTML=coverage.html
MIGRATE_BIN=migrate
MIGRATIONS_DIR=db/migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable



migrate-up:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-status:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# Сброс в случае ошибок (ОСТОРОЖНО: принудительно откатывает и ставит миграции заново)
migrate-force-reset:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up




test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=$(TEST_COVERAGE_FILE) ./...
	go tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)



lint:
	golangci-lint run



run:
	go run main.go

clean:
	rm -f $(APP_NAME) $(TEST_COVERAGE_FILE) $(TEST_COVERAGE_HTML)
