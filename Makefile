.PHONY: install

# go install determines the install directory itself (GOBIN, else GOPATH/bin),
# and the build cache makes the old find-based staleness check unnecessary.
install:
	go install ./cmd/makex
