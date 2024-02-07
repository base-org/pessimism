APP_NAME = pessimism

LINTER_VERSION = v1.52.1
LINTER_URL = https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh

GET_LINT_CMD = "curl -sSfL $(LINTER_URL) | sh -s -- -b $(go env GOPATH)/bin $(LINTER_VERSION)"

RED = \033[0;34m
GREEN = \033[0;32m
BLUE = \033[0;34m
COLOR_END = \033[0;39m

TEST_LIMIT = 500s

build-app:
	@echo "$(BLUE)» building application binary... $(COLOR_END)"
	@CGO_ENABLED=0 go build -a -tags netgo -o bin/$(APP_NAME) ./cmd/
	@echo "Binary successfully built"

run-app:
	@./bin/${APP_NAME}

.PHONY: go-gen-mocks
go-gen-mocks:
	@echo "generating go mocks..."
	@GO111MODULE=on go generate --run "mockgen*" ./...
	
.PHONY: test
test:
	@go test ./internal/... -timeout $(TEST_LIMIT)

.PHONY: test-e2e
e2e-test:
	@docker compose up -d
	@go test ./e2e/...  -timeout $(TEST_LIMIT) -deploy-config ../.devnet/devnetL1.json -parallel=4 -v
	@docker compose down

.PHONY: lint
lint:
	@echo "$(GREEN) Linting repository Go code...$(COLOR_END)"
	@if ! command -v golangci-lint &> /dev/null; \
	then \
    	echo "golangci-lint command could not be found...."; \
		echo "\nTo install, please run $(GREEN)  $(GET_LINT_CMD) $(COLOR_END)"; \
		echo "\nBuild instructions can be found at: https://golangci-lint.run/usage/install/."; \
    	exit 1; \
	fi

	@golangci-lint run

gosec:
	@echo "$(GREEN) Running security scan with gosec...$(COLOR_END)"
	gosec ./...

.PHONY: docker-build
docker-build:
	@echo "$(GREEN) Building docker image...$(COLOR_END)"
	@docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	@echo "$(GREEN) Running docker image...$(COLOR_END)"
	@ docker run -p 8080:8080 -p 7300:7300 --env-file=config.env -it -v ${PWD}/genesis.json:/app/genesis.json $(APP_NAME)

.PHONY: metrics-docs
metrics-docs: build-app
		@echo "$(GREEN) Generating metric documentation...$(COLOR_END)"
		@./bin/pessimism doc metrics


devnet-allocs:
	@echo "$(GREEN) Generating devnet allocs...$(COLOR_END)"
	@./scripts/devnet-allocs.sh
