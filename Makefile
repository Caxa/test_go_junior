.PHONY: test test-cover lint run build clean

# Переменные
APP_NAME=go-people-api
TEST_COVERAGE_FILE=coverage.out
TEST_COVERAGE_HTML=coverage.html

# Запуск тестов
test:
	go test -v -race ./...

# Запуск тестов с покрытием
test-cover:
	go test -v -race -coverprofile=$(TEST_COVERAGE_FILE) ./...
	go tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)

# Проверка линтером
lint:
	golangci-lint run

# Запуск приложения
run:
	go run main.go

# Сборка приложения
build:
	go build -o $(APP_NAME) main.go

# Очистка
clean:
	rm -f $(APP_NAME) $(TEST_COVERAGE_FILE) $(TEST_COVERAGE_HTML)
