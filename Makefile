APP_NAME = pessimism
LINT_VERSION = v1.52.1

build-app:
	@echo "\033[0;34m» building application binary... \033[0;39m"
	@CGO_ENABLED=0 go build -a -tags netgo -o bin/$(APP_NAME) ./cmd/pessimism/
	@echo "Binary successfully built"

run-app: 
	@./bin/${APP_NAME}

.PHONY: test
test:
	@ go test ./...

.PHONY: lint
lint:
	@echo "\033[0;32m» Linting repository Go code...\033[0;39m"
	@if ! command -v golangci-lint &> /dev/null; \
	then \
    	echo "golangci-lint command could not be found...."; \
		echo "To install, please run \033[0;34m curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.52.1 \033[0;39m"; \
    	exit 1; \
	fi

	@golangci-lint run