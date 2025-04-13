.PHONY: all deps build install clean

all: build

deps:
	go mod download

build:
	mkdir -p target
	go build -o target/onkyo cmd/cli/main.go cmd/cli/chat.go

install: target/onkyo
	cp target/onkyo ~/.local/bin/

clean:
	rm -rf target