.PHONY: help lint mock test clean

help: ## Показать справку
	@echo "Доступные команды:"
	@echo "  make lint       - Запустить golangci-lint"
	@echo "  make mock       - Сгенерировать моки с помощью mockery"
	@echo "  make test       - Запустить тесты"
	@echo "  make clean      - Удалить сгенерированные моки"
	@echo "  make all        - Запустить lint и mock"

lint: ## Запустить golangci-lint
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

mock: ## Сгенерировать моки
	mockery

test: ## Запустить тесты
	go test -v ./...

all: lint mock ## Запустить lint и сгенерировать моки

