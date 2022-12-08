GOPATH=$(shell go env GOPATH)
GOLANGCI_LINT=$(GOPATH)/bin/golangci-lint

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@echo "==> Linting codebase"
	@$(GOLANGCI_LINT) run

.PHONY: test
test:
	@echo "==> Running tests"
	GO111MODULE=on go test -v

.PHONY: test-cover
test-cover:
	@echo "==> Running Tests with coverage"
	GO111MODULE=on go test -cover .

.PHONY: fuzz
fuzz:
	@echo "==> Fuzz testing"
	go test -fuzz .

$(GOLANGCI_LINT):
	# Install golangci-lint. The configuration for it is in the .golangci.yml
	# file in the root of the repository
	echo ${GOPATH}
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.50.1

$(GOFUZZBUILD):
	cd / && go get -u github.com/dvyukov/go-fuzz/go-fuzz-build

$(GOFUZZ):
	cd / && go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-dep
