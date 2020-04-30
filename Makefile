export GOBIN ?= $(CURDIR)/bin

GOLINT = $(GOBIN)/golint
STATICCHECK = $(GOBIN)/staticcheck

.PHONY: test
test:
	go test -race -v ./...

.PHONY: lint
lint: $(GOLINT) $(STATICCHECK)
	$(GOLINT) ./...
	$(STATICCHECK) ./...

$(GOLINT):
	go install golang.org/x/lint/golint

$(STATICCHECK):
	go install honnef.co/go/tools/cmd/staticcheck
