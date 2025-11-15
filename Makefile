.PHONY: help lint mock test integration-tests test-all clean

help: ## Показать справку
	@echo "Доступные команды:"
	@echo "  make lint              - Запустить golangci-lint"
	@echo "  make mock              - Сгенерировать моки с помощью mockery"
	@echo "  make test              - Запустить unit тесты"
	@echo "  make integration-tests - Запустить интеграционные тесты"
	@echo "  make test-all          - Запустить все тесты (unit + integration)"
	@echo "  make clean             - Удалить сгенерированные моки"
	@echo "  make all               - Запустить lint и mock"

lint: ## Запустить golangci-lint
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

mock: ## Сгенерировать моки
	mockery

test: ## Запустить unit тесты
	go test -v ./... -short

integration-tests: ## Запустить интеграционные тесты
	@powershell -Command "$$env:INTEGRATION_TESTS='1'; go test -v ./internal/integration/..."

test-all: ## Запустить все тесты (unit + integration)
	@powershell -Command "$$env:INTEGRATION_TESTS='1'; go test -v ./..."

all: lint mock ## Запустить lint и сгенерировать моки

