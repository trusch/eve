all: bin docker

bin: eve.amd64 eve.arm eve-ctl.amd64 eve-ctl.arm

docker: eve.docker eve-ctl.docker

eve.amd64: $(shell find ./ -name "*.go")
	GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags -static' -o eve.amd64 github.com/trusch/eve
eve.arm: $(shell find ./ -name "*.go")
	GOOS=linux GOARCH=arm go build -o eve.arm github.com/trusch/eve
eve-ctl.amd64: $(shell find ./ -name "*.go")
	GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags -static' -o eve-ctl.amd64 github.com/trusch/eve/eve-ctl
eve-ctl.arm: $(shell find ./ -name "*.go")
	GOOS=linux GOARCH=arm go build -o eve-ctl.arm github.com/trusch/eve/eve-ctl

eve.docker: eve.amd64
	docker build -t trusch/eve:latest -f eve.dockerfile .
eve-ctl.docker: eve-ctl.amd64
	docker build -t trusch/eve-ctl:latest -f eve-ctl.dockerfile .

clean:
	-rm eve.amd64 eve.arm eve-ctl.amd64 eve-ctl.arm
