BINDIR=_out

GOLANGCI_LINT_VERSION=1.54.2
GOLANGCI_LINT_BIN=$(BINDIR)/golangci-lint
GOLANGCI_LINT_VERSION_TAG=v${GOLANGCI_LINT_VERSION}

all: deployer

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -rf _out

.PHONY: clean-deps
clean-deps:
	rm -rf vendor

.PHONY: update-deps
update-deps:
	go mod tidy && go mod vendor

.PHONY: update-manifests
update-manifests:
	./pkg/manifests/yaml/update.sh

.PHONY: update-version
update-version:
	@mkdir -p pkg/version || :
	@hack/make-version.sh > pkg/version/version.go

deployer-static: outdir
	CGO_ENABLED=0 go build -o _out/deployer ./cmd/deployer

deployer: outdir update-version
	go build -o _out/deployer ./cmd/deployer/

outdir:
	@mkdir -p _out || :

.PHONY: release-manifests
release-manifests-k8s: deployer
	@_out/deployer -P kubernetes:v1.24 render > _out/deployer-manifests-allinone.yaml

.PHONY: test-unit
test-unit:
	go test ./pkg/...

.PHONY: test-unit-cover
test-unit-cover:
	go test -coverprofile=coverage.out ./pkg/...

.PHONY: test-e2e-positive
test-e2e-positive: build-e2e
	_out/e2e.test -ginkgo.focus='\[PositiveFlow\]'

.PHONY: test-e2e-negative
test-e2e-negative: build-e2e
	_out/e2e.test -ginkgo.focus='\[NegativeFlow\]'

.PHONY: gofmt
gofmt:
	@echo "Running gofmt"
	gofmt -s -w `find . -path ./vendor -prune -o -type f -name '*.go' -print`

# this is meant for developers only.
# DO NOT WIRE THIS IN CI! Let's use https://golangci-lint.run/usage/install/#github-actions instead
.PHONY: dev-lint
dev-lint: _out/golangci-lint
	$(GOLANGCI_LINT_BIN) run

.PHONY: build-e2e
build-e2e: _out/e2e.test

_out/e2e.test: outdir test/e2e/*.go
	go test -v -c -o _out/e2e.test ./test/e2e/

_out/golangci-lint: outdir
	@if [ ! -x "$(GOLANGCI_LINT_BIN)" ]; then\
		echo "Downloading golangci-lint $(GOLANGCI_LINT_VERSION)";\
		curl -JL https://github.com/golangci/golangci-lint/releases/download/$(GOLANGCI_LINT_VERSION_TAG)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz -o _out/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz;\
		tar xz -C _out -f _out/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz;\
		ln -sf golangci-lint-1.54.2-linux-amd64/golangci-lint _out/golangci-lint;\
	else\
		echo "Using golangci-lint cached at $(GOLANGCI_LINT_BIN)";\
	fi

