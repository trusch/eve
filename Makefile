all: bin docker aci

bin: bin/eve.amd64 bin/eve.arm bin/eve-ctl.amd64 bin/eve-ctl.arm

docker: eve.docker eve-ctl.docker

aci: bin/eve-latest-amd64.aci bin/eve-latest-armv7l.aci bin/eve-ctl-latest-amd64.aci bin/eve-ctl-latest-armv7l.aci

bin/eve.amd64: $(shell find ./ -name "*.go")
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags -static' -o bin/eve.amd64 github.com/trusch/eve
bin/eve.arm: $(shell find ./ -name "*.go")
	mkdir -p bin
	GOOS=linux GOARCH=arm go build -o bin/eve.arm github.com/trusch/eve
bin/eve-ctl.amd64: $(shell find ./ -name "*.go")
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags -static' -o bin/eve-ctl.amd64 github.com/trusch/eve/eve-ctl
bin/eve-ctl.arm: $(shell find ./ -name "*.go")
	mkdir -p bin
	GOOS=linux GOARCH=arm go build -o bin/eve-ctl.arm github.com/trusch/eve/eve-ctl

eve.docker: bin/eve.amd64
	docker build -t trusch/eve:latest -f scripts/eve.dockerfile .
eve-ctl.docker: bin/eve-ctl.amd64
	docker build -t trusch/eve-ctl:latest -f scripts/eve-ctl.dockerfile .

bin/eve-latest-amd64.aci: bin/eve.amd64 scripts/build-aci-eve-amd64.sh
	bash scripts/build-aci-eve-amd64.sh
bin/eve-latest-armv7l.aci: bin/eve.arm scripts/build-aci-eve-armv7l.sh
	bash scripts/build-aci-eve-armv7l.sh
bin/eve-ctl-latest-amd64.aci: bin/eve.amd64 scripts/build-aci-eve-ctl-amd64.sh
	bash scripts/build-aci-eve-ctl-amd64.sh
bin/eve-ctl-latest-armv7l.aci: bin/eve.arm scripts/build-aci-eve-ctl-armv7l.sh
	bash scripts/build-aci-eve-ctl-armv7l.sh

clean:
	-rm -rf bin
