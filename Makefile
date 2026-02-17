# Detect OS
OS := $(shell uname)

.PHONY: build run clean test test-short test-coverage test-integration test-integration-short test-all coverage-check lint lint-fix vet deps compress compress-all release release-all

BINARY=moko
VERSION=0.3.0
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
BUILDFLAGS=-trimpath
CGO=CGO_ENABLED=0
TEST_RESULTS_DIR=test-results
DIST_DIR=dist

build:
	@mkdir -p $(DIST_DIR)
	$(CGO) go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY) ./cmd/moko

run: build
	./$(DIST_DIR)/$(BINARY)

clean:
	@echo "Cleaning build artifacts and test results..."
	@rm -rf $(DIST_DIR) $(TEST_RESULTS_DIR)
	@rm -f $(BINARY) $(BINARY)-*
	@rm -f coverage.out coverage*.out coverage.html
	@rm -f *.test
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@find . -name "*~" -delete 2>/dev/null || true
	@echo "✓ Repository cleaned"

test:
	@mkdir -p $(TEST_RESULTS_DIR)
	go test -v ./... 2>&1 | tee $(TEST_RESULTS_DIR)/test-output.txt

test-short:
	@mkdir -p $(TEST_RESULTS_DIR)
	go test -v -short ./... 2>&1 | tee $(TEST_RESULTS_DIR)/test-short-output.txt

test-coverage:
	@mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race -coverprofile=$(TEST_RESULTS_DIR)/coverage.out ./... 2>&1 | tee $(TEST_RESULTS_DIR)/test-coverage-output.txt
	go tool cover -html=$(TEST_RESULTS_DIR)/coverage.out -o $(TEST_RESULTS_DIR)/coverage.html
	@echo ""
	@echo "Coverage report saved to $(TEST_RESULTS_DIR)/coverage.html"
	@echo ""
	@go tool cover -func=$(TEST_RESULTS_DIR)/coverage.out | grep total

test-integration:
	@mkdir -p $(TEST_RESULTS_DIR)
	go test -v -tags=integration ./... 2>&1 | tee $(TEST_RESULTS_DIR)/test-integration-output.txt

test-integration-short:
	@mkdir -p $(TEST_RESULTS_DIR)
	go test -v -tags=integration -short ./... 2>&1 | tee $(TEST_RESULTS_DIR)/test-integration-short-output.txt

test-all: test-short test-integration-short

coverage-check:
	@mkdir -p $(TEST_RESULTS_DIR)
	@go test -coverprofile=$(TEST_RESULTS_DIR)/coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=$(TEST_RESULTS_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 40" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below 40% threshold"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets threshold"; \
	fi

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

lint-fix:
	@echo "Running linters with auto-fix..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run --fix ./...

vet:
	go vet ./...

deps:
	go mod tidy

# Cross-compilation targets
build-linux:
	@mkdir -p $(DIST_DIR)
	$(CGO) GOOS=linux GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-linux-amd64 ./cmd/moko

build-darwin:
	@mkdir -p $(DIST_DIR)
	$(CGO) GOOS=darwin GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-amd64 ./cmd/moko
	$(CGO) GOOS=darwin GOARCH=arm64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-arm64 ./cmd/moko

build-windows:
	@mkdir -p $(DIST_DIR)
	$(CGO) GOOS=windows GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-windows-amd64.exe ./cmd/moko

build-all: build-linux build-darwin build-windows

# UPX compression (requires upx: brew install upx)
# Note: UPX is not compatible with macOS (security/code signing issues)
# macOS binaries are already optimized with -s -w -trimpath
compress:
ifeq ($(OS),Darwin)
	@echo "UPX not supported on macOS. Binary already optimized with -s -w -trimpath."
else
	upx --best --lzma $(DIST_DIR)/$(BINARY)
endif

compress-all:
	upx --best --lzma $(DIST_DIR)/$(BINARY)-linux-amd64
	upx --best --lzma $(DIST_DIR)/$(BINARY)-windows-amd64.exe
	@echo "Skipping UPX for macOS binaries (not supported)"

# Build and compress
release: build compress

release-all: build-all compress-all
