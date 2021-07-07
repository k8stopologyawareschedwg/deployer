all: deployer

.PHONY: clan
clean:
	rm -rf _out

.PHONY: update-deps
update-deps:
	go mod tidy && go mod vendor

.PHONY: update-manifests
update-manifests:
	./pkg/manifests/manifests/update.sh

deployer: outdir
	go build -o _out/deployer ./cmd/deployer/

outdir:
	mkdir -p _out

