all: deployer

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
update-version: prepare
	@hack/make-version.sh > pkg/version/version.go

.PHONY: update-images
update-images: hacks
	@_out/updateimage \
		< pkg/images/images.json \
		> pkg/images/consts.go && \
	gofmt -w pkg/images/consts.go

deployer-static: prepare
	CGO_ENABLED=0 go build -o _out/deployer ./cmd/deployer

deployer: prepare update-version
	go build -o _out/deployer ./cmd/deployer/

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
build-e2e: _out/rte-e2e.test

_out/rte-e2e.test: prepare test/e2e/*.go
	go test -v -c -o _out/e2e.test ./test/e2e/

.PHONY: hacks
hacks: prepare
	go build -o _out/mirrorurl hack/tools/mirrorurl/main.go
	go build -o _out/mirrorgen hack/tools/mirrorgen/main.go
	go build -o _out/updateimage hack/tools/updateimage/main.go

.PHONY: prepare
prepare: outdir
	@mkdir -p pkg/version || :

outdir:
	@mkdir -p _out || :

