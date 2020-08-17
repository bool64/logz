GOLANGCI_LINT_VERSION := "v1.30.0"
GO ?= go
# detecting GOPATH and removing trailing "/" if any
GOPATH = $(realpath $(shell go env GOPATH))
export GO111MODULE = on

lint:
	@test -s $(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION) || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp $(GOLANGCI_LINT_VERSION) && mv /tmp/golangci-lint $(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION))
	@$(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION) run ./... --fix

bench:
	@$(GO) test -bench=. -count=10 -run=^a  ./... | tee /dev/tty >bench-$(shell git symbolic-ref HEAD --short | tr / - 2>/dev/null).txt
	@test -s $(GOPATH)/bin/benchstat || bash -c 'cd /tmp;GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) get -u golang.org/x/perf/cmd/benchstat'
	@benchstat bench-$(shell git symbolic-ref HEAD --short | tr / - 2>/dev/null).txt


