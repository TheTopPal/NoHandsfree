BINARY = bin/nohandsfree-cli.exe

.PHONY: build test lint clean install uninstall

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

install: build
	./$(BINARY) install

uninstall:
	./$(BINARY) uninstall
