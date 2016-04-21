install: ${GOBIN}/makex

${GOBIN}/makex: $(shell /usr/bin/find -type f -and -name '*.go')
	go install ./cmd/makex
