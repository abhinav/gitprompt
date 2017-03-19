PACKAGES = $(shell glide nv)

.PHONY: test
test:
	go test -race -v $(PACKAGES)
