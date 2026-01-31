.PHONY: build clean install test run-fetcher run-telegram deps migrate

# Build variables
BINARY_DIR=bin
MAIL_FETCHER=mail-fetcher
TELEGRAM_SERVICE=telegram-service

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
BUILD_FLAGS=-v -ldflags="-s -w"

all: build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build binaries
build: deps
	@echo "Building binaries..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(MAIL_FETCHER) ./cmd/mail-fetcher
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(TELEGRAM_SERVICE) ./cmd/telegram-service
	@echo "Build complete!"

# Build for production with optimizations
build-prod: deps
	@echo "Building production binaries..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(MAIL_FETCHER) ./cmd/mail-fetcher
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(TELEGRAM_SERVICE) ./cmd/telegram-service
	@echo "Production build complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)

# Run mail-fetcher locally
run-fetcher: build
	@echo "Running mail-fetcher..."
	./$(BINARY_DIR)/$(MAIL_FETCHER) -config configs/config.yaml

# Run telegram-service locally
run-telegram: build
	@echo "Running telegram-service..."
	./$(BINARY_DIR)/$(TELEGRAM_SERVICE) -config configs/config.yaml

# Install to system (requires root)
install: build
	@echo "Installing to system..."
	sudo ./scripts/install.sh

# Setup systemd services (requires root)
setup-services:
	@echo "Setting up systemd services..."
	sudo ./scripts/setup-services.sh

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@if [ -z "$(DB_PASSWORD)" ]; then \
		echo "Error: DB_PASSWORD not set"; \
		exit 1; \
	fi
	@mysql -u mail_user -p$(DB_PASSWORD) mail_to_tg < migrations/001_initial_schema.sql
	@mysql -u mail_user -p$(DB_PASSWORD) mail_to_tg < migrations/002_add_indexes.sql
	@echo "Migrations complete!"

# Create database and user
create-db:
	@echo "Creating database and user..."
	@if [ -z "$(DB_ROOT_PASSWORD)" ]; then \
		echo "Error: DB_ROOT_PASSWORD not set"; \
		exit 1; \
	fi
	@if [ -z "$(DB_PASSWORD)" ]; then \
		echo "Error: DB_PASSWORD not set"; \
		exit 1; \
	fi
	@mysql -u root -p$(DB_ROOT_PASSWORD) -e "CREATE DATABASE IF NOT EXISTS mail_to_tg CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	@mysql -u root -p$(DB_ROOT_PASSWORD) -e "CREATE USER IF NOT EXISTS 'mail_user'@'localhost' IDENTIFIED BY '$(DB_PASSWORD)';"
	@mysql -u root -p$(DB_ROOT_PASSWORD) -e "GRANT ALL PRIVILEGES ON mail_to_tg.* TO 'mail_user'@'localhost';"
	@mysql -u root -p$(DB_ROOT_PASSWORD) -e "FLUSH PRIVILEGES;"
	@echo "Database created!"

# Generate encryption key
gen-key:
	@echo "Generating encryption key..."
	@openssl rand -base64 32

# Backup
backup:
	@echo "Running backup..."
	sudo ./scripts/backup.sh

# Start services
start:
	sudo systemctl start mail-fetcher telegram-service

# Stop services
stop:
	sudo systemctl stop mail-fetcher telegram-service

# Restart services
restart:
	sudo systemctl restart mail-fetcher telegram-service

# View logs
logs-fetcher:
	sudo journalctl -u mail-fetcher -f

logs-telegram:
	sudo journalctl -u telegram-service -f

# Status
status:
	sudo systemctl status mail-fetcher telegram-service

# Help
help:
	@echo "Mail-to-Telegram Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  deps          - Install Go dependencies"
	@echo "  build         - Build binaries"
	@echo "  build-prod    - Build production binaries (static, optimized)"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-fetcher   - Run mail-fetcher locally"
	@echo "  run-telegram  - Run telegram-service locally"
	@echo "  install       - Install to system (requires root)"
	@echo "  setup-services- Setup systemd services (requires root)"
	@echo "  create-db     - Create database and user (requires DB_ROOT_PASSWORD and DB_PASSWORD)"
	@echo "  migrate       - Run database migrations (requires DB_PASSWORD)"
	@echo "  gen-key       - Generate encryption key"
	@echo "  backup        - Run backup script"
	@echo "  start         - Start services"
	@echo "  stop          - Stop services"
	@echo "  restart       - Restart services"
	@echo "  logs-fetcher  - View mail-fetcher logs"
	@echo "  logs-telegram - View telegram-service logs"
	@echo "  status        - Show service status"
	@echo "  help          - Show this help"
