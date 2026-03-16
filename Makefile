GOBUILD = go build
GOTEST = go test

# tools
goimports_version = v0.43.0
yamlfmt_version = v0.21.0
golangci_version = v2.11.3
govulncheck_version = v1.1.4
actionlint_version = v1.7.11
ghalint_version = v1.5.5
pinact_version = v3.9.0

# targets
cwallet:
	$(GOBUILD) -o ./cli/build/cwallet cli/main.go

install:
	@cp ./cli/build/cwallet /usr/bin/

test:
	$(GOTEST) -short ./...

testcov:
	$(GOTEST) -short ./... -coverprofile coverage.out

opencov:
	go tool cover -html coverage.out

# format and lint
format: format-go format-yaml

format-go:
	go run golang.org/x/tools/cmd/goimports@${goimports_version} -format-only -w ./

format-yaml:
	go run github.com/google/yamlfmt/cmd/yamlfmt@${yamlfmt_version}

lint: lint-go lint-actions-all

lint-fix: lint-go-fix lint-actions-all-fix

lint-go:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${golangci_version} run

lint-go-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${golangci_version} run --fix

lint-actions-all: lint-actionlint lint-ghalint

lint-actions-all-fix: pinact lint-actions-all

pinact:
	go run github.com/suzuki-shunsuke/pinact/v3/cmd/pinact@${pinact_version} run

lint-actionlint:
	go run github.com/rhysd/actionlint/cmd/actionlint@${actionlint_version}

lint-ghalint:
	go run github.com/suzuki-shunsuke/ghalint/cmd/ghalint@${ghalint_version} run
