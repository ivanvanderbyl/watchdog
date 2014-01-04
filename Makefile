DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html

deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d
	go get github.com/axw/gocov/gocov
	go get -u github.com/matm/gocov-html
	github.com/hashicorp/serf/testutil

test: deps
	go list ./... | xargs -n1 go test

