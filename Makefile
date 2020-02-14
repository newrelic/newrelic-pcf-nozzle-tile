# Don't assume PATH settings
export PATH := $(PATH):$(GOPATH)/bin
INTEGRATION  := newrelic-pcf-nozzle
BINARY_NAME   = nr-fh-nozzle
GO_FILES     := ./...
GO_INTEGRATION_FILE := ./tests/...
#Release version must be mayor.minor.patch for tile generator
RELEASE_TAG   ?= 0.0.1
TEST_DEPS     = github.com/axw/gocov/gocov github.com/AlekSi/gocov-xml

all: release

build: clean deps test-deps compile test integration-test

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: removing binaries and coverage file..."
	@rm -rfv dist product release coverage.xml

deps:
	@echo "=== $(INTEGRATION) === [ deps ]: downloading dependencies..."
	@dep ensure

test-deps:
	@echo "=== $(INTEGRATION) === [ test-deps ]: installing testing dependencies..."
	@go get -v $(TEST_DEPS)

test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go clean -testcache
	@gocov test $(GO_FILES) | gocov-xml > coverage.xml

integration-test: compile
	@echo "=== $(INTEGRATION) === [ integration test ]: running integration tests..."
	@go clean -testcache
	@go test $(GO_INTEGRATION_FILE) -tags=integration -v

compile:
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@mkdir -p dist 
	@go build -ldflags "-X main.Version=$(RELEASE_TAG)" -o dist/$(BINARY_NAME)

compile-linux: clean deps
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@mkdir -p dist 
	@env GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version=$(RELEASE_TAG)" -o dist/$(BINARY_NAME)

release: build compile-linux
	@echo "=== $(INTEGRATION) === [ release ]: generating release..."
	@tile build $(RELEASE_TAG)

push: compile-linux
	@echo "=== $(INTEGRATION) === [ push ]: pushing to test environment..."
	@cf login -a $(CF_API_URL) --skip-ssl-validation -u $(CF_USER) -p $(CF_PASSWORD) -o nr-firehose-nozzle-org
	@cf push

.PHONY: all build clean compile test-deps test release integration-test push compile-linux