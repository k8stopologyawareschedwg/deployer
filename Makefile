all: deployer

.PHONY: vet
vet:
	go vet ./...

.PHONY: clan
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

.PHONY: test-unit
test-unit:
	go test ./pkg/...

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

.PHONY: build-e2e
build-e2e: _out/e2e.test

_out/e2e.test: outdir test/e2e/*.go
	go test -v -c -o _out/e2e.test ./test/e2e/

