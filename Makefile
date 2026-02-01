# DivvyDoo Makefile

# Variables

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;36m
NC=\033[0m # No Color

help:
	@echo -e "$(BLUE)DivvyDoo Makefile Commands:$(NC)"
	@echo -e "  $(GREEN)make build$(NC)        - Build the DivvyDoo project"
	@echo -e "  $(GREEN)make test$(NC)         - Run tests for DivvyDoo"
	@echo -e "  $(GREEN)make clean$(NC)        - Clean build artifacts"
	@echo -e "  $(GREEN)make docs$(NC)         - Generate documentation"
	@echo -e "  $(GREEN)make help$(NC)         - Show this help message"

install:
	@echo -e "$(YELLOW)Installing backend dependencies...$(NC)"
	go mod download

build: 
	@echo -e "$(YELLOW)Building backend...$(NC)"
	go build -o ../bin/divvydoo-backend main.go

dev:
	@echo -e "$(YELLOW)Running backend...$(NC)"
	go run cmd/api/main.go