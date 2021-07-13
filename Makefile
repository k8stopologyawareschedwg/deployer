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
	./pkg/manifests/manifests/update.sh

deployer: outdir
	go build -o _out/deployer ./cmd/deployer/

outdir:
	mkdir -p _out || :

.PHONY: test-unit
test-unit:
	go test ./pkg/...

.PHONY: gofmt
gofmt:
	@echo "Running gofmt"
	gofmt -s -w `find . -path ./vendor -prune -o -type f -name '*.go' -print`
