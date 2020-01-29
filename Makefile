# Don't assume PATH settings
export PATH := $(PATH):$(GOPATH)/bin
INTEGRATION  := newrelic-pcf-nozzle
BINARY_NAME   = nr-fh-nozzle
GO_FILES     := ./...
#Release version must be mayor.minor.patch for tile generator
RELEASE_TAG   := 0.0.1
TEST_DEPS     = github.com/axw/gocov/gocov github.com/AlekSi/gocov-xml

all: release
build: clean deps test-deps test compile
clean:
	@echo "=== $(INTEGRATION) === [ clean ]: removing binaries and coverage file..."
	@rm -rfv dist product release coverage.xml
deps:
	@echo "=== $(INTEGRATION) === [ deps ]: downloading dependencies..."
	@dep ensure
compile:
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@mkdir -p dist 
	@go build -ldflags "-X main.Version=$(RELEASE_TAG)" -o dist/$(BINARY_NAME)
test-deps:
	@echo "=== $(INTEGRATION) === [ test-deps ]: installing testing dependencies..."
	@go get -v $(TEST_DEPS)
test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@gocov test $(GO_FILES) | gocov-xml > coverage.xml
release: build
	@echo "=== $(INTEGRATION) === [ release ]: generating release..."
	@tar -czvf dist/$(BINARY_NAME).tar.gz dist/$(BINARY_NAME)
	@tile build $(RELEASE_TAG)

.PHONY: all build clean compile test-deps test release