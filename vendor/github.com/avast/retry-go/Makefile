SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=
DEP?=$$(which dep)
VERSION?=$$(cat VERSION)
LINTER?=$$(which golangci-lint)
LINTER_VERSION=1.15.0

ifeq ($(OS),Windows_NT)
	DEP_VERS=dep-windows-amd64
	LINTER_FILE=golangci-lint-$(LINTER_VERSION)-windows-amd64.zip
	LINTER_UNPACK= >| app.zip; unzip -j app.zip -d $$GOPATH/bin; rm app.zip
else ifeq ($(OS), Darwin)
	LINTER_FILE=golangci-lint-$(LINTER_VERSION)-darwin-amd64.tar.gz
	LINTER_UNPACK= | tar xzf - -C $$GOPATH/bin --wildcards --strip 1 "**/golangci-lint"
else
	DEP_VERS=dep-linux-amd64
	LINTER_FILE=golangci-lint-$(LINTER_VERSION)-linux-amd64.tar.gz
	LINTER_UNPACK= | tar xzf - -C $$GOPATH/bin --wildcards --strip 1 "**/golangci-lint"
endif

setup:
	go get -u github.com/pierrre/gotestcover
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/robertkrimen/godocdown/godocdown
	@if [ "$(LINTER)" = "" ]; then\
		curl -L https://github.com/golangci/golangci-lint/releases/download/v$(LINTER_VERSION)/$(LINTER_FILE) $(LINTER_UNPACK) ;\
		chmod +x $$GOPATH/bin/golangci-lint;\
	fi
	@if [ "$(DEP)" = "" ]; then\
		curl -L https://github.com/golang/dep/releases/download/v0.3.1/$(DEP_VERS) >| $$GOPATH/bin/dep;\
		chmod +x $$GOPATH/bin/dep;\
	fi
	dep ensure

generate: ## Generate README.md
	godocdown >| README.md

test: generate test_and_cover_report lint

test_and_cover_report:
	gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m

cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

lint: ## Run all the linters
	golangci-lint run

ci: test_and_cover_report ## Run all the tests but no linters - use https://golangci.com integration instead

build:
	go build

release: ## Release new version
	git tag | grep -q $(VERSION) && echo This version was released! Increase VERSION! || git tag $(VERSION) && git push origin $(VERSION) && git tag v$(VERSION) && git push origin v$(VERSION)

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
