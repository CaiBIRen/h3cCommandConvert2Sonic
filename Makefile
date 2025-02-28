#
#LDFLAGS := "-s -w -X main.buildTime=$(shell date -u '+%Y-%m-%dT%I:%M:%S%p')"
LDFLAGS := "-X main.buildTime=$(shell date -u '+%Y-%m-%dT%I:%M:%S%p')  -X main.version=${NET_AGENT_TAG}"

BIN_NAME=sonic-unis-framework

build:
	go build -ldflags $(LDFLAGS) -o $(BIN_NAME) main/main.go
race:
	go build -ldflags $(LDFLAGS) -o $(BIN_NAME) -race main/main.go
swag:
	swag init -g main/main.go
	go build -ldflags $(LDFLAGS) -o $(BIN_NAME) main/main.go
clean:
	rm -f $(BIN_NAME)
install:
	mkdir -p ./debian/sonic-unis-framework/usr/local/bin
	cp sonic-unis-framework ./debian/sonic-unis-framework/usr/local/bin/
# 	sudo install -m 755 sonic-unis-framework /usr/local/bin
# .PHONY: release